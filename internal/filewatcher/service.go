// Package filewatcher provides a service that watches config files and issues
// a debounced event with changes
package filewatcher

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fiffeek/hyprwhenthen/internal/config"
	"github.com/fiffeek/hyprwhenthen/internal/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Callback interface {
	OnEvent(context.Context) error
}

type Service struct {
	cfg       *config.Config
	watcher   *fsnotify.Watcher
	debouncer *utils.Debouncer
	callback  Callback
}

func NewService(cfg *config.Config, callback Callback) *Service {
	return &Service{
		cfg:       cfg,
		debouncer: utils.NewDebouncer(),
		callback:  callback,
	}
}

func (s *Service) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Context cancelled for filewatcher, shutting down")
		return context.Cause(ctx)
	})

	logrus.Debug("Starting watcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("cant create watcher: %w", err)
	}
	s.watcher = watcher

	eg.Go(func() error {
		return s.debouncer.Run(ctx)
	})

	eg.Go(func() error {
		<-ctx.Done()
		s.debouncer.Cancel()
		logrus.Debug("Context cancelled, shutting watcher down")
		if err := watcher.Close(); err != nil {
			logrus.WithError(err).Error("Cant close watcher on exit")
		}
		return context.Cause(ctx)
	})

	eg.Go(func() error {
		logrus.Debug("Initialized watcher")
		if err := s.runServiceLoop(ctx, watcher); err != nil {
			return fmt.Errorf("cant run service loop: %w", err)
		}
		logrus.Debug("Exiting watcher")
		return nil
	})

	if err := watcher.Add(s.cfg.Get().Dir); err != nil {
		return fmt.Errorf("cant watch config file changes: %w", err)
	}

	return eg.Wait()
}

func (s *Service) runServiceLoop(ctx context.Context, watcher *fsnotify.Watcher) error {
	logrus.Debug("Starting filewatcher goroutine")
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return errors.New("watcher channel is closed")
			}

			logrus.WithFields(logrus.Fields{
				"name":      event.Name,
				"operation": event.Op,
			}).Debug("Received filewatcher event")

			s.debouncer.Do(ctx, time.Duration(*s.cfg.Get().General.HotReloadDebounceTimer), s.callback.OnEvent)
			logrus.WithFields(logrus.Fields{"fun": s.callback.OnEvent}).Debug("Scheduled debounced update")
		case err, ok := <-watcher.Errors:
			if !ok {
				return errors.New("watcher error channel is closed")
			}
			if err != nil {
				return fmt.Errorf("watcher error received: %w", err)
			}
		case <-ctx.Done():
			logrus.Debug("Context cancelled, shutting fswatcher down")
			return context.Cause(ctx)
		}
	}
}
