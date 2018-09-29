package store

import (
	"log"

	"github.com/yodo-io/cryptogen/pkg/crypto/worker"
)

// UpdateJobs listens to job status feed, updates task status in database
func UpdateJobs(s Provider, feed <-chan worker.JobUpdate) {
	go func() {
		for u := range feed {
			if u.Error != nil {
				log.Printf("Job %s encountered an error: %v", u.JobID, u.Error)
			} else {
				log.Printf("Job %s changed to %s", u.JobID, u.Status)
			}
			status := JobStatus{
				Status:      u.Status,
				SecretPaths: u.SecretPaths,
			}
			if err := s.SetStatus(u.JobID, status); err != nil {
				log.Printf("Failed to update status for job %s: %v", u.JobID, err)
			}
		}
	}()
}
