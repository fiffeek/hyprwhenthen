package workerpool

import (
	"os"
	"time"

	"github.com/fiffeek/hyprwhenthen/internal/utils"
	"github.com/google/uuid"
)

type Result struct {
	JobID uuid.UUID
	Err   error
	Exec  string
}

type Job struct {
	ID         uuid.UUID
	RoutingKey string
	extraEnv   map[string]string
	Exec       string
	Timeout    *time.Duration
}

func NewJob(env map[string]string, exec string, timeout *time.Duration, maybeRoutingKey *string) *Job {
	routingKey := maybeRoutingKey
	jobID := uuid.New()
	if routingKey == nil {
		routingKey = utils.JustPtr(jobID.String())
	} else {
		routingKey = utils.JustPtr(os.Expand(*routingKey, func(key string) string {
			value, ok := env[key]
			if !ok {
				return os.Getenv(key)
			}
			return value
		}))
	}
	return &Job{
		extraEnv:   env,
		Exec:       exec,
		ID:         jobID,
		Timeout:    timeout,
		RoutingKey: *routingKey,
	}
}
