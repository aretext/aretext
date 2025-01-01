package protocol

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"syscall"
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

	var oob []byte
	if startSessionMsg, ok := msg.(*StartSessionMsg); ok {
		if startSessionMsg.Tty == nil {
			return errors.New("StartSessionMsg.Tty must not be nil")
		}
		oob = syscall.UnixRights(int(startSessionMsg.Tty.Fd()))
	}

	_, _, err = conn.WriteMsgUnix(data, oob, nil)
	if err != nil {
		return fmt.Errorf("net.WriteMsgUnix: %w", err)
	}

	return nil
}

// ReceiveMessage receives a mesage over a Unix socket.
func ReceiveMessage(conn *net.UnixConn) (Message, error) {
	var oob [128]byte
	var headerData [4]byte

	n, oobn, _, _, err := conn.ReadMsgUnix(headerData[:], oob[:])
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

		if oobn == 0 {
			return nil, errors.New("Missing expected OOB data in StartSessionMsg")
		}

		cmsgs, err := syscall.ParseSocketControlMessage(oob[0:oobn])
		if err != nil {
			return nil, fmt.Errorf("syscall.ParseSocketControlMessage: %w", err)
		}

		fds, err := syscall.ParseUnixRights(&cmsgs[0])
		if err != nil {
			return nil, fmt.Errorf("syscall.ParseUnixRights: %w", err)
		} else if len(fds) != 1 {
			return nil, errors.New("invalid number of file descriptors received for Tty")
		}

		ttyFile := os.NewFile(uintptr(fds[0]), "")
		if ttyFile == nil {
			return nil, errors.New("invalid file descriptor for Tty")
		}

		msg.Tty = ttyFile

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
