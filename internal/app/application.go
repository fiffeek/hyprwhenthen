package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/fiffeek/hyprwhenthen/internal/config"
	"github.com/fiffeek/hyprwhenthen/internal/eventprocessor"
	"github.com/fiffeek/hyprwhenthen/internal/filewatcher"
	"github.com/fiffeek/hyprwhenthen/internal/hypr"
	"github.com/fiffeek/hyprwhenthen/internal/signal"
	"github.com/fiffeek/hyprwhenthen/internal/workerpool"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Application struct {
	cfg            *config.Config
	hypr           *hypr.Service
	pool           *workerpool.Service
	eventProcessor *eventprocessor.Service
	startOnce      sync.Once
	signalHandler  *signal.Handler
	watcher        *filewatcher.Service
}

func NewApplication(ctx context.Context, cancelCause context.CancelCauseFunc, configPath string, workersNum int, queueSize int) (*Application, error) {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("cant load config: %w", err)
	}

	watcher := filewatcher.NewService(cfg, cfg)

	hypr, err := hypr.NewService(ctx, cfg)
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

	handler := signal.NewHandler(cancelCause)

	return &Application{
		cfg:            cfg,
		hypr:           hypr,
		pool:           pool,
		eventProcessor: processor,
		signalHandler:  handler,
		watcher:        watcher,
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
		{Fun: a.hypr.Run, Name: "hypr"},
		{Fun: a.pool.Run, Name: "bg worker pool"},
		{Fun: a.eventProcessor.Run, Name: "event processor"},
		{Fun: a.signalHandler.Run, Name: "signal handler"},
		{Fun: a.watcher.Run, Name: "watch config changes"},
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
