package protocol

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"syscall"
)

// TODO
func SendClientHello(ctx context.Context, conn *net.UnixConn, msg *ClientHelloMsg, pts *os.File) error {
	ptsFd := int(pts.Fd())
	return sendFramedMsgData(ctx, conn, msg, &ptsFd)
}

func ReceiveClientHello(ctx context.Context, conn *net.UnixConn) (*ClientHelloMsg, *os.File, error) {

}

func SendClientGoodbye(ctx context.Context, conn *net.UnixConn, msg *ClientGoodbyeMsg) error {
	return nil
}

func ReceiveClientGoodbye(ctx context.Context, conn *net.UnixConn) (*ClientGoodbyeMsg, error) {
	return nil, nil
}

func SendServerHello(ctx context.Context, conn *net.UnixConn, msg *ServerHelloMsg) error {
	return nil
}

func ReceiveServerHello(ctx context.Context, conn *net.UnixConn) (*ServerHelloMsg, error) {
	return nil, nil
}

func SendServerGoodbye(ctx context.Context, conn *net.UnixConn, msg *ServerGoodbyeMsg) error {
	return nil
}

func ReceiveServerGoodbye(ctx context.Context, conn *net.UnixConn) (*ServerGoodbyeMsg, error) {
	return nil, nil
}

func SendTerminalResize(ctx context.Context, conn *net.UnixConn, msg *TerminalResizeMsg) error {
	return nil
}

func ReceiveTerminalResize(ctx context.Context, conn *net.UnixConn) (*TerminalResizeMsg, error) {
	return nil, nil
}

func sendFramedMsgData[M any](ctx context.Context, conn *net.UnixConn, msg *M, oobFileDesc *int) error {
	if msg == nil {
		return errors.New("Message cannot be nil")
	}

	encodedMsg, err := json.Marshal(*msg)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	if len(encodedMsg) >= math.MaxUint16 {
		return errors.New("Message too long")
	}

	data := make([]byte, len(encodedMsg)+4)
	binary.BigEndian.PutUint16(data[0:], uint16(len(encodedMsg)))
	copy(data[4:], encodedMsg)

	var oob []byte
	if oobFileDesc != nil {
		oob = syscall.UnixRights(*oobFileDesc)
	}

	_, _, err = conn.WriteMsgUnix(data, oob, nil)
	if err != nil {
		return fmt.Errorf("net.WriteMsgUnix: %w", err)
	}

	return nil
}

func receiveFramedMsgData[M any](ctx context.Context, conn *net.UnixConn) (*M, *int, error) {
}
