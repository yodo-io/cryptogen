package store

import (
	"fmt"
	"strings"

	"github.com/go-redis/redis"
)

// RedisConf is the config struct for redis clients
type RedisConf struct {
	Address  string
	Password string
	DB       int
}

// NewRedis creates new storage client backed by Redis.
func NewRedis(conf RedisConf) *RedisStore {
	rc := redis.NewClient(&redis.Options{
		Addr:     conf.Address,
		Password: conf.Password,
		DB:       conf.DB,
	})
	return &RedisStore{client: rc}

}

// RedisStore is a storage client backed by Redis. Unless Redis is somehow replicated and/or persisted,
// this is mostly useful for development or testing purposes.
type RedisStore struct {
	client *redis.Client
}

func taskKey(taskID, suffix string) string {
	return fmt.Sprintf("cryptogen.task.%s.%s", taskID, suffix)
}

// we don't really do anything with the paths in Redis for now, other than
// storing and retrieving. So a string blob will do for
func serializePaths(paths []string) string {
	return strings.Join(paths, ":")
}

func deserializePaths(paths string) []string {
	if len(paths) == 0 {
		return []string{}
	}
	return strings.Split(paths, ":")
}

// Ping does a ping to test connection, returns an error on failure
func (r *RedisStore) Ping() error {
	if _, err := r.client.Ping().Result(); err != nil {
		return err
	}
	return nil
}

// SetStatus stores the status for given task in Redis using a key generated based on taskID
func (r *RedisStore) SetStatus(taskID string, status JobStatus) error {
	skey := taskKey(taskID, "status")
	pkey := taskKey(taskID, "paths")

	if err := r.client.Set(skey, status.Status, 0).Err(); err != nil {
		return err
	}
	if err := r.client.Set(pkey, serializePaths(status.SecretPaths), 0).Err(); err != nil {
		return err
	}
	return nil
}

// GetStatus retrieves the status for given task in Redis using a key generated based on taskID
// It returns ErrNotFound if nothing was found for the requested key
func (r *RedisStore) GetStatus(taskID string) (JobStatus, error) {
	var out JobStatus

	skey := taskKey(taskID, "status")
	pkey := taskKey(taskID, "paths")

	// get status
	status, err := r.client.Get(skey).Result()
	if err != nil {
		return out, err
	}
	if status == "" {
		return out, ErrNotFound
	}

	// get paths, if any
	paths, err := r.client.Get(pkey).Result()
	if err != nil {
		return out, err
	}

	out = JobStatus{
		Status:      status,
		SecretPaths: deserializePaths(paths),
	}
	return out, nil
}
