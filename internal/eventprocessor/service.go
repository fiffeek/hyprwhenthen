package eventprocessor

import (
	"context"
	"errors"
	"fmt"
	"hyprwhenthen/internal/config"
	"hyprwhenthen/internal/hypr"
	"hyprwhenthen/internal/workerpool"
	"regexp"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	ipc       *hypr.IPC
	pool      *workerpool.Service
	cfg       *config.Config
	startOnce sync.Once
}

func NewService(ipc *hypr.IPC, pool *workerpool.Service, cfg *config.Config) (*Service, error) {
	return &Service{
		ipc:  ipc,
		pool: pool,
		cfg:  cfg,
	}, nil
}

func (s *Service) Run(ctx context.Context) error {
	var err error
	s.startOnce.Do(func() {
		err = s.run(ctx)
	})
	return err
}

func (s *Service) run(ctx context.Context) error {
	hyprEventsChannel := s.ipc.Listen()
	resultsChannel := s.pool.Listen()
	defer s.pool.Stop()
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		for {
			select {
			case result, ok := <-resultsChannel:
				if !ok {
					return errors.New("results channel closed")
				}
				logrus.WithError(result.Err).WithFields(logrus.Fields{"id": result.JobID, "exec": result.Exec}).Info("Worker result collected")

			case <-ctx.Done():
				logrus.Debug("Event processor context cancelled, shutting down")
				return context.Cause(ctx)
			}
		}
	})

	eg.Go(func() error {
		for {
			select {
			case event, ok := <-hyprEventsChannel:
				if !ok {
					return errors.New("hypr events channel closed")
				}
				logrus.WithFields(logrus.Fields{"type": event.EventType, "context": event.EventContext}).Debug("Hypr event received")
				if err := s.process(ctx, event); err != nil {
					return fmt.Errorf("dispatch unsuccessful: %w", err)
				}

			case <-ctx.Done():
				logrus.Debug("Event processor context cancelled, shutting down")
				return context.Cause(ctx)
			}
		}
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("goroutines for service failed %w", err)
	}
	return nil
}

func (s *Service) process(ctx context.Context, event *hypr.Event) error {
	cfg := s.cfg.Get()
	onEvents, found := cfg.OnEvents[event.EventType]
	if !found {
		logrus.Debugf("System is not configured to react to %s event type", event.EventType)
		return nil
	}

	for _, matcher := range onEvents {
		reg, err := regexp.Compile(matcher.When)
		if err != nil {
			return fmt.Errorf("cant compile regexp %s: %w", matcher.When, err)
		}

		if !reg.Match([]byte(event.EventContext)) {
			logrus.WithFields(logrus.Fields{"event": event.EventContext, "regex": matcher.When}).Debug("Event does not match regex, skipping...")
			continue
		}

		env := map[string]string{}
		matches := reg.FindStringSubmatch(event.EventContext)
		for i, match := range matches {
			env["REGEX_GROUP_"+strconv.Itoa(i)] = match
		}

		job := workerpool.NewJob(env, matcher.Then, matcher.Timeout)
		logrus.WithFields(logrus.Fields{"id": job.ID, "exec": job.Exec}).Info("Submitting execution to the pool")
		if err := s.pool.Submit(ctx, job); err != nil {
			return fmt.Errorf("cant submit a job for execution: %w", err)
		}
		logrus.Debug("Submission successful")
	}
	return nil
}
