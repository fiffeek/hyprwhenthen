// Package dial provides unix socket helpers.
package dial

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/sirupsen/logrus"
)

func GetUnixSocketConnection(ctx context.Context, socketPath string) (net.Conn, func(), error) {
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("socket not found at %s", socketPath)
	}

	d := &net.Dialer{}
	conn, err := d.DialContext(ctx, "unix", socketPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to socket: %w", err)
	}

	return conn, func() {
		if err := conn.Close(); err != nil {
			logrus.WithError(err).Debug("Failed to close connection")
		}
	}, nil
}
