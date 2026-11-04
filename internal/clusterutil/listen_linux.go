//go:build linux

package clusterutil

import (
	"context"
	"net"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// Listen uses SO_REUSEPORT for cluster workers. Node distributes shared
// server connections from the primary; Gode's worker processes do not pass
// native listener handles, so the kernel is used as the shared acceptor for
// the same port. Non-cluster listeners keep the standard behavior.
func Listen(network, address string) (net.Listener, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return net.Listen(network, address)
	}
	if os.Getenv("GODE_CLUSTER_WORKER_ID") == "" && os.Getenv("NODE_UNIQUE_ID") == "" {
		return net.Listen(network, address)
	}
	lc := net.ListenConfig{Control: func(_, _ string, c syscall.RawConn) error {
		var sockErr error
		if err := c.Control(func(fd uintptr) {
			sockErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
		}); err != nil {
			return err
		}
		return sockErr
	}}
	return lc.Listen(context.Background(), network, address)
}
