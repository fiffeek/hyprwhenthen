package workerpool

import (
	"context"
	"errors"
	"fmt"
	"hyprwhenthen/internal/config"
	"os"
	"os/exec"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	workers   int
	jobs      chan *Job
	cfg       *config.Config
	results   chan *Result
	closed    chan struct{}
	startOnce sync.Once
	closeOnce sync.Once
}

func NewService(workersNum int, queueSize int, cfg *config.Config) (*Service, error) {
	if workersNum <= 0 {
		return nil, errors.New("workersNum has to be > 0")
	}
	if queueSize < 0 {
		return nil, errors.New("queue must be >= 0")
	}
	return &Service{
		workers: workersNum,
		jobs:    make(chan *Job, queueSize),
		closed:  make(chan struct{}),
		results: make(chan *Result, queueSize),
		cfg:     cfg,
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
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case s.jobs <- job:
		return nil
	}
}

func (s *Service) Stop() {
	s.closeOnce.Do(func() {
		close(s.closed)
		close(s.jobs)
	})
}

func (s *Service) Run(ctx context.Context) error {
	var err error
	s.startOnce.Do(func() {
		eg, ctx := errgroup.WithContext(ctx)
		for i := 0; i < s.workers; i++ {
			eg.Go(func() error { return s.runJob(ctx) })
		}
		err = eg.Wait()
		close(s.results)
	})
	return err
}

func (s *Service) runJob(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("worker panic: %v", r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case job, ok := <-s.jobs:
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
