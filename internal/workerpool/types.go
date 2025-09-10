package workerpool

import (
	"time"

	"github.com/google/uuid"
)

type Result struct {
	JobID uuid.UUID
	Err   error
	Exec  string
}

type Job struct {
	ID       uuid.UUID
	extraEnv map[string]string
	Exec     string
	Timeout  *time.Duration
}

func NewJob(env map[string]string, exec string, timeout *time.Duration) *Job {
	return &Job{
		extraEnv: env,
		Exec:     exec,
		ID:       uuid.New(),
		Timeout:  timeout,
	}
}
