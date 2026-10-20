//go:build darwin

package listener

import (
	"syscall"

	"github.com/sagernet/sing/common/control"

	"golang.org/x/sys/unix"
)

// shareUIDControl sets SO_REUSESHAREUID on TCP listeners.
//
// XNU's in_pcbbind rejects a non-root bind to an address and port while any
// socket owned by a different uid still holds it, including the TIME_WAIT and
// FIN_WAIT_2 connections left behind after the owning process has exited, and
// SO_REUSEADDR does not lift this check. Accepted connections inherit the flag,
// so when a listener running as root goes away, the same address can be bound
// again by an unprivileged process right away instead of after the connections
// time out. Live listeners still conflict as before.
func shareUIDControl() control.Func {
	return func(network, address string, conn syscall.RawConn) error {
		return control.Raw(conn, func(fd uintptr) error {
			return unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSESHAREUID, 1)
		})
	}
}
