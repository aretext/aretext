package protocol

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
)

const maxMsgLen = 102400 // 100 KiB

// SendMessage sends a message over a Unix socket.
func SendMessage(conn *net.UnixConn, msg Message) error {
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
	binary.BigEndian.PutUint16(data[0:], uint16(msgType))
	binary.BigEndian.PutUint16(data[2:], uint16(len(encodedMsg)))
	copy(data[4:], encodedMsg)

	_, _, err = conn.WriteMsgUnix(data, nil, nil)
	if err != nil {
		return fmt.Errorf("net.WriteMsgUnix: %w", err)
	}

	return nil
}

// ReceiveMessage receives a mesage over a Unix socket.
func ReceiveMessage(conn *net.UnixConn) (Message, error) {
	var headerData [4]byte

	n, _, _, _, err := conn.ReadMsgUnix(headerData[:], nil)
	if err != nil {
		return nil, fmt.Errorf("net.ReadMsgUnix: %w", err)
	} else if n != 4 {
		return nil, errors.New("Too few bytes read for message header")
	}

	msgType := msgType(binary.BigEndian.Uint16(headerData[0:]))
	msgLen := int(binary.BigEndian.Uint16(headerData[2:]))

	if msgLen > maxMsgLen {
		return nil, errors.New("Invalid length for message")
	}

	msgData := make([]byte, msgLen)
	n, _, _, _, err = conn.ReadMsgUnix(msgData, nil)
	if err != nil {
		return nil, fmt.Errorf("net.ReadMsgUnix: %w", err)
	} else if n != msgLen {
		return nil, errors.New("Too few bytes read for message data")
	}

	switch msgType {
	case startSessionMsgType:
		var msg StartSessionMsg
		if err := json.Unmarshal(msgData, &msg); err != nil {
			return nil, fmt.Errorf("json.Unmarshal: %w", err)
		}
		return &msg, nil

	case resizeTerminalMsgType:
		var msg ResizeTerminalMsg
		if err := json.Unmarshal(msgData, &msg); err != nil {
			return nil, fmt.Errorf("json.Unmarshal: %w", err)
		}
		return &msg, nil

	default:
		return nil, errors.New("Unrecognized message type")
	}
}
