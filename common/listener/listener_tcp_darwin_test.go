//go:build darwin

package listener

import (
	"context"
	"net"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

func shareUID(t *testing.T, conn syscall.Conn) int {
	t.Helper()
	rawConn, err := conn.SyscallConn()
	if err != nil {
		t.Fatal(err)
	}
	var value int
	var sockErr error
	err = rawConn.Control(func(fd uintptr) {
		value, sockErr = unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSESHAREUID)
	})
	if err != nil {
		t.Fatal(err)
	}
	if sockErr != nil {
		t.Fatal(sockErr)
	}
	return value
}

func TestShareUIDControlIsInheritedByAcceptedConnections(t *testing.T) {
	listenConfig := net.ListenConfig{Control: shareUIDControl()}
	listener, err := listenConfig.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	if shareUID(t, listener.(*net.TCPListener)) == 0 {
		t.Fatal("SO_REUSESHAREUID not set on listener")
	}
	client, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	accepted, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer accepted.Close()
	if shareUID(t, accepted.(*net.TCPConn)) == 0 {
		t.Fatal("SO_REUSESHAREUID not inherited by accepted connection")
	}
}

func TestShareUIDControlKeepsLiveListenerExclusive(t *testing.T) {
	listenConfig := net.ListenConfig{Control: shareUIDControl()}
	listener, err := listenConfig.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	second, err := listenConfig.Listen(context.Background(), "tcp", listener.Addr().String())
	if err == nil {
		second.Close()
		t.Fatal("second listener bound to an address that is still listening")
	}
}
