package hypr

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"hyprwhenthen/internal/config"
	"hyprwhenthen/internal/dial"
	"hyprwhenthen/internal/utils"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	instanceSignature string
	xdgRuntimeDir     string
	events            chan *Event
	cfg               *config.Config
}

func NewService(ctx context.Context, cfg *config.Config) (*Service, error) {
	signature := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if signature == "" {
		return nil, errors.New("HYPRLAND_INSTANCE_SIGNATURE environment variable not set - are you running under Hyprland?")
	}

	xdgRuntimeDir, err := utils.GetXDGRuntimeDir()
	if err != nil {
		return nil, fmt.Errorf("cant get xdg runtime dir: %w", err)
	}

	return &Service{
		instanceSignature: signature,
		xdgRuntimeDir:     xdgRuntimeDir,
		events:            make(chan *Event, 100),
		cfg:               cfg,
	}, nil
}

func (i *Service) Listen() <-chan *Event {
	return i.events
}

func (i *Service) Run(ctx context.Context) error {
	socketPath := GetHyprEventsSocket(i.xdgRuntimeDir, i.instanceSignature)
	eg, ctx := errgroup.WithContext(ctx)

	conn, connTeardown, err := dial.GetUnixSocketConnection(ctx, socketPath)
	if err != nil {
		return fmt.Errorf("cant open unix events socket connection to %s: %w", socketPath, err)
	}

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Hypr IPC context cancelled, closing connection to unblock scanner")
		connTeardown()
		return context.Cause(ctx)
	})

	eg.Go(func() error {
		defer close(i.events)
		defer connTeardown()

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				logrus.Debug("Hypr IPC context cancelled during processing")
				return context.Cause(ctx)
			default:
			}

			line := scanner.Text()
			cfg := i.cfg.Get()
			found, event := getRegisteredEvent(cfg, line)
			if !found {
				logrus.WithFields(logrus.Fields{"line": line}).Debug("Event not registered")
				continue
			}

			select {
			case i.events <- event:
				logrus.Debug("Monitors event sent")
			case <-ctx.Done():
				logrus.Debug("Hypr IPC context cancelled during event send")
				return context.Cause(ctx)
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scanner error: %w", err)
		}

		logrus.Debug("Hypr IPC scanner finished")
		return nil
	})

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("goroutines for hypr ipc failed %w", err)
	}
	return nil
}
