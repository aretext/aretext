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

func SendClientHello(ctx context.Context, conn *net.UnixConn, msg *ClientHelloMsg, pts *os.File) error {
	return sendFramedMsgData(ctx, conn, msg, pts)
}

func SendClientGoodbye(ctx context.Context, conn *net.UnixConn, msg *ClientGoodbyeMsg) error {
	return sendFramedMsgData(ctx, conn, msg, nil)
}

func SendServerHello(ctx context.Context, conn *net.UnixConn, msg *ServerHelloMsg) error {
	return sendFramedMsgData(ctx, conn, msg, nil)
}

func SendServerGoodbye(ctx context.Context, conn *net.UnixConn, msg *ServerGoodbyeMsg) error {
	return sendFramedMsgData(ctx, conn, msg, nil)
}

func SendTerminalResize(ctx context.Context, conn *net.UnixConn, msg *TerminalResizeMsg) error {
	return sendFramedMsgData(ctx, conn, msg, nil)
}

func ReceiveMessage(ctx context.Context, conn *net.UnixConn) (

func sendFramedMsgData[M Message](ctx context.Context, conn *net.UnixConn, msg *M, oobFile *os.File) error {
	if msg == nil {
		return errors.New("Message cannot be nil")
	}

	msgType := msgTypeForMessage(msg)
	if msgType == invalidMsgType {
		return errors.New("Invalid message type")
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
	binary.BigEndian.PutUint16(data[4:], uint16(msgType))
	copy(data[8:], encodedMsg)

	var oob []byte
	if oobFile != nil {
		oob = syscall.UnixRights(int(oobFile.Fd()))
	}

	_, _, err = conn.WriteMsgUnix(data, oob, nil)
	if err != nil {
		return fmt.Errorf("net.WriteMsgUnix: %w", err)
	}

	return nil
}

func receiveFramedMsgData[M any](ctx context.Context, conn *net.UnixConn) (*M, *os.File, error) {
}
