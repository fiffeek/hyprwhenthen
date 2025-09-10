package app

import (
	"context"
	"fmt"
	"hyprwhenthen/internal/config"
	"hyprwhenthen/internal/eventprocessor"
	"hyprwhenthen/internal/hypr"
	"hyprwhenthen/internal/workerpool"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Application struct {
	cfg            *config.Config
	hypr           *hypr.IPC
	pool           *workerpool.Service
	eventProcessor *eventprocessor.Service
	startOnce      sync.Once
}

func NewApplication(ctx context.Context, configPath string, workersNum int, queueSize int) (*Application, error) {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("cant load config: %w", err)
	}

	hypr, err := hypr.NewIPC(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("cant init hypr: %w", err)
	}

	pool, err := workerpool.NewService(workersNum, queueSize, cfg)
	if err != nil {
		return nil, fmt.Errorf("cant init pool: %w", err)
	}

	processor, err := eventprocessor.NewService(hypr, pool, cfg)
	if err != nil {
		return nil, fmt.Errorf("cant init event processor: %w", err)
	}

	return &Application{
		cfg:            cfg,
		hypr:           hypr,
		pool:           pool,
		eventProcessor: processor,
	}, nil
}

func (a *Application) Run(ctx context.Context) error {
	var err error
	a.startOnce.Do(func() {
		err = a.run(ctx)
	})
	return err
}

func (a *Application) run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	backgroundGoroutines := []struct {
		Fun  func(context.Context) error
		Name string
	}{
		{Fun: a.hypr.RunEventLoop, Name: "hypr ipc"},
		{Fun: a.pool.Run, Name: "power detector dbus"},
		{Fun: a.eventProcessor.Run, Name: "reloader"},
	}
	for _, bg := range backgroundGoroutines {
		eg.Go(func() error {
			fields := logrus.Fields{"name": bg.Name}
			logrus.WithFields(fields).Debug("Starting")
			if err := bg.Fun(ctx); err != nil {
				logrus.WithFields(fields).Debug("Exited with error")
				return fmt.Errorf("%s failed: %w", bg.Name, err)
			}
			logrus.WithFields(fields).Debug("Finished")
			return nil
		})
	}

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Context cancelled, shutting down")
		return context.Cause(ctx)
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("app run eg failed: %w", err)
	}

	logrus.Info("Shutdown complete")
	return nil
}
