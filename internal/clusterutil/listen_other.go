//go:build !linux

package clusterutil

import "net"

// Listen falls back to a normal listener on platforms without SO_REUSEPORT.
func Listen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}
