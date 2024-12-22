package protocol

import (
	"context"
	"net"
	"os"
)

// TODO
func SendClientHello(ctx context.Context, conn *net.UnixConn, m *ClientHelloMsg, pts *os.File) error {
	return nil
}

func ReceiveClientHello(ctx context.Context, conn *net.UnixConn) (*ClientHelloMsg, *os.File, error) {
	return nil, nil, nil
}

func SendClientGoodbye(ctx context.Context, conn *net.UnixConn, m *ClientGoodbyeMsg) error {
	return nil
}

func ReceiveClientGoodbye(ctx context.Context, conn *net.UnixConn) (*ClientGoodbyeMsg, error) {
	return nil, nil
}

func SendServerHello(ctx context.Context, conn *net.UnixConn, m *ServerHelloMsg) error {
	return nil
}

func ReceiveServerHello(ctx context.Context, conn *net.UnixConn) (*ServerHelloMsg, error) {
	return nil, nil
}

func SendServerGoodbye(ctx context.Context, conn *net.UnixConn, m *ServerGoodbyeMsg) error {
	return nil
}

func ReceiveServerGoodbye(ctx context.Context, conn *net.UnixConn) (*ServerGoodbyeMsg, error) {
	return nil, nil
}

func SendTerminalResize(ctx context.Context, conn *net.UnixConn, m *TerminalResizeMsg) error {
	return nil
}

func ReceiveTerminalResize(ctx context.Context, conn *net.UnixConn) (*TerminalResizeMsg, error) {
	return nil, nil
}
