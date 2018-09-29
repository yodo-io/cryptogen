package worker

import (
	"fmt"
	"math/rand"
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
	Status string
	Error  error
	JobID  string
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
			if err := w.work(j); err != nil {
				w.Feed <- JobUpdate{JobID: j.ID, Error: err, Status: JobStatusError}
				continue
			}
			w.Feed <- JobUpdate{JobID: j.ID, Status: JobStatusComplete}
		}
	}()
}

func (w *Worker) work(j Job) error {
	res, err := w.crypto.GenerateAssets(j.ID, j.Req)
	if err != nil {
		return err
	}
	prefix := "/secret/crypto/" + j.ID
	if err := w.kms.StoreAssets(prefix, res); err != nil {
		return err
	}
	return nil
}
