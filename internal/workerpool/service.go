// Package workerpool provides a service that executes jobs in background.
package workerpool

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"sync"

	"github.com/fiffeek/hyprwhenthen/internal/config"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	workers      int
	workerQueues []chan *Job
	cfg          *config.Config
	results      chan *Result
	closed       chan struct{}
	startOnce    sync.Once
	closeOnce    sync.Once
}

func NewService(workersNum, queueSize int, cfg *config.Config) (*Service, error) {
	if workersNum <= 0 {
		return nil, errors.New("workersNum has to be > 0")
	}
	if queueSize < 0 {
		return nil, errors.New("queue must be >= 0")
	}

	workerQueues := make([]chan *Job, workersNum)
	for i := range workersNum {
		workerQueues[i] = make(chan *Job, queueSize)
	}

	return &Service{
		workers:      workersNum,
		workerQueues: workerQueues,
		closed:       make(chan struct{}),
		results:      make(chan *Result, queueSize*workersNum),
		cfg:          cfg,
	}, nil
}

func (s *Service) Listen() <-chan *Result {
	return s.results
}

func (s *Service) Submit(ctx context.Context, job *Job) error {
	select {
	case <-s.closed:
		return errors.New("pool is closed")
	default:
	}

	logrus.WithFields(logrus.Fields{"id": job.ID, "routing_key": job.RoutingKey}).Debug("Queuing a job")
	workerIndex, err := s.hashRoutingKey(job.RoutingKey)
	if err != nil {
		return fmt.Errorf("cant calculate worker index: %w", err)
	}
	workerIndex %= len(s.workerQueues)
	logrus.WithFields(logrus.Fields{"id": job.ID, "worker": workerIndex}).Debug("Assigned the job to a worker")

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case s.workerQueues[workerIndex] <- job:
		return nil
	}
}

func (s *Service) Stop() {
	s.closeOnce.Do(func() {
		close(s.closed)
		for _, queue := range s.workerQueues {
			close(queue)
		}
	})
}

func (s *Service) Run(ctx context.Context) error {
	var err error
	s.startOnce.Do(func() {
		eg, ctx := errgroup.WithContext(ctx)
		for i := 0; i < s.workers; i++ {
			eg.Go(func() error { return s.runWorker(ctx, i) })
		}
		err = eg.Wait()
		close(s.results)
	})
	return err
}

func (s *Service) runWorker(ctx context.Context, workerID int) error {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("worker %d panic: %v", workerID, r)
		}
	}()

	jobQueue := s.workerQueues[workerID]
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case job, ok := <-jobQueue:
			if !ok {
				return nil
			}
			err := s.executeJob(ctx, job)
			select {
			case s.results <- &Result{JobID: job.ID, Err: err, Exec: job.Exec}:
			case <-ctx.Done():
				return context.Cause(ctx)
			}
		}
	}
}

func (s *Service) executeJob(ctx context.Context, job *Job) error {
	timeout := job.Timeout
	if timeout == nil {
		timeout = s.cfg.Get().General.Timeout
	}

	jobCtx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()
	// nolint: gosec
	cmd := exec.CommandContext(jobCtx, "bash", "-c", job.Exec)
	env := append([]string{}, os.Environ()...)
	for key, value := range job.extraEnv {
		env = append(env, key+"="+value)
	}
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	logrus.Debugf("Command %s output %s", job.Exec, string(out))
	if err != nil {
		if jobCtx.Err() != nil {
			return context.Cause(jobCtx)
		}
		return fmt.Errorf("job %s errored: %w", job.ID, err)
	}

	return nil
}

func (s *Service) hashRoutingKey(routingKey string) (int, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(routingKey))
	return int(h.Sum32()), err
}
