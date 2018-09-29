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

type Worker struct {
	Feed   chan JobUpdate
	queue  chan Job
	kms    kms.Provider
	crypto crypto.Provider
}

type Config struct {
	Kms    kms.Provider
	Crypto crypto.Provider
}

type JobUpdate struct {
	Status      string
	Error       error
	JobID       string
	SecretPaths []string // paths to generated secrets
}

type Job struct {
	ID  string
	Req crypto.GenerateCryptoRequest
}

func (w *Worker) NewJob(req crypto.GenerateCryptoRequest) Job {
	jobID := fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Intn(100000))
	job := Job{
		ID:  jobID,
		Req: req,
	}
	w.queue <- job
	return job
}

func New(c *Config) *Worker {
	w := &Worker{
		Feed:   make(chan JobUpdate),
		queue:  make(chan Job),
		kms:    c.Kms,
		crypto: c.Crypto,
	}
	for i := 0; i < 3; i++ {
		w.worker()
	}
	return w
}

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
		if err := w.kms.StoreAssets(p, a.Secrets); err != nil {
			return nil, err
		}
	}
	return paths, nil
}
