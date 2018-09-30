package worker

import (
	"fmt"
	"math/rand"
	"path"
	"time"

	"github.com/yodo-io/cryptogen/pkg/crypto"
	"github.com/yodo-io/cryptogen/pkg/kms"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// List of task statuses
const (
	JobStatusError      string = "new"
	JobStatusProcessing string = "processing"
	JobStatusComplete   string = "complete"
)

// Worker type, read jobs from a queue, generate crypto assets and write them to a KMS
// Job status updates are published to a channel
type Worker struct {
	Feed   chan JobUpdate
	queue  chan Job
	kms    kms.Provider
	crypto crypto.Provider
}

// Config struct for worker, holding its dependencies
type Config struct {
	Kms    kms.Provider
	Crypto crypto.Provider
}

// JobUpdate is published whenever a job changes it's status. Current job status
// will be in the `Status` field. Any errors are in the `Error` field.
type JobUpdate struct {
	Status      string
	Error       error
	JobID       string
	SecretPaths []string // paths to generated secrets
}

// Job to process by the worker
type Job struct {
	ID  string
	Req crypto.GenerateCryptoRequest
}

// NewJob creates a new job with a randomised ID and enqueues it in the worker
func (w *Worker) NewJob(req crypto.GenerateCryptoRequest) Job {
	jobID := fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Intn(100000))
	job := Job{
		ID:  jobID,
		Req: req,
	}
	w.queue <- job
	return job
}

// New creates a new worker from given Config
func New(c *Config) *Worker {
	w := &Worker{
		Feed:   make(chan JobUpdate),
		queue:  make(chan Job),
		kms:    c.Kms,
		crypto: c.Crypto,
	}
	// FIXME: don't hardcode number of workers
	for i := 0; i < 3; i++ {
		w.worker()
	}
	return w
}

// worker routing. process jobs as they come in, publish updates
func (w *Worker) worker() {
	go func() {
		for j := range w.queue {
			w.Feed <- JobUpdate{JobID: j.ID, Status: JobStatusProcessing}
			paths, err := w.work(j)
			if err != nil {
				w.Feed <- JobUpdate{JobID: j.ID, Error: err, Status: JobStatusError}
				continue
			}
			w.Feed <- JobUpdate{
				Status:      JobStatusComplete,
				JobID:       j.ID,
				SecretPaths: paths,
			}
		}
	}()
}

// does the actual work of generating crypto assets
func (w *Worker) work(j Job) ([]string, error) {
	assets, err := w.crypto.GenerateAssets(j.ID, j.Req)
	if err != nil {
		return nil, err
	}

	paths := make([]string, len(assets))
	prefix := "/secret/cryptogen"
	for i, a := range assets {
		p := path.Join(prefix, a.Path)
		paths[i] = p
		if err := w.kms.Store(p, a.Entries); err != nil {
			return nil, err
		}
	}
	return paths, nil
}
