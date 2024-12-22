package protocol

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"syscall"
)

// SendMessage sends a message over a Unix socket between a client and a server.
func SendMessage(ctx context.Context, conn *net.UnixConn, msg Message) error {
	if msg == nil {
		return errors.New("Message cannot be nil")
	}

	msgType := msgTypeForMessage(msg)
	if msgType == invalidMsgType {
		return errors.New("Invalid message type")
	}

	encodedMsg, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	if len(encodedMsg) >= math.MaxUint16 {
		return errors.New("Message too long")
	}

	data := make([]byte, len(encodedMsg)+4)
	binary.BigEndian.PutUint16(data[4:], uint16(msgType))
	binary.BigEndian.PutUint16(data[0:], uint16(len(encodedMsg)))
	copy(data[8:], encodedMsg)

	var oob []byte
	if clientHelloMsg, ok := msg.(*ClientHelloMsg); ok {
		oob = syscall.UnixRights(int(clientHelloMsg.Pts.Fd()))
	}

	_, _, err = conn.WriteMsgUnix(data, oob, nil)
	if err != nil {
		return fmt.Errorf("net.WriteMsgUnix: %w", err)
	}

	return nil
}

// ReceiveMessage receives a mesage over a Unix socket.
func ReceiveMessage(ctx context.Context, conn *net.UnixConn) (Message, error) {
	// TODO
	return nil, nil
}
