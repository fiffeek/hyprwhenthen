// Package signal provides signal handling functionality.
package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Interrupted struct{ sig os.Signal }

func (i Interrupted) Error() string { return "interrupted by " + i.sig.String() }

func (i Interrupted) ExitCode() int {
	if s, ok := i.sig.(syscall.Signal); ok {
		return 128 + int(s)
	}
	return 1
}

type Handler struct {
	sigChan     chan os.Signal
	cancelCause context.CancelCauseFunc
}

func NewHandler(cancelCause context.CancelCauseFunc) *Handler {
	return &Handler{
		sigChan:     make(chan os.Signal, 1),
		cancelCause: cancelCause,
	}
}

func (h *Handler) Run(ctx context.Context) error {
	signal.Notify(h.sigChan, syscall.SIGTERM, syscall.SIGINT)
	logrus.Debug("Signal notifications registered for SIGTERM, SIGINT")

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Signal handler goroutine exiting")
		signal.Stop(h.sigChan)
		return nil
	})

	eg.Go(func() error {
		logrus.Debug("Signal handler goroutine started")
		defer close(h.sigChan)
		for {
			select {
			case sig := <-h.sigChan:
				logrus.WithField("signal", sig).Debug("Signal received")
				switch sig {
				case syscall.SIGTERM, syscall.SIGINT:
					logrus.WithField("signal", sig).Info("Received termination signal, shutting down gracefully")
					h.cancelCause(&Interrupted{sig})
					return &Interrupted{sig}
				}
			case <-ctx.Done():
				logrus.Debug("Signal handler context done, exiting")
				return context.Cause(ctx)
			}
		}
	})

	return eg.Wait()
}
