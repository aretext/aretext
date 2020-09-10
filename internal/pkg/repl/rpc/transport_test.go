package rpc

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testClient struct {
	conn io.ReadWriteCloser
}

func (c *testClient) Send(header RequestHeader, data []byte) error {
	headerData, err := json.Marshal(&header)
	if err != nil {
		return err
	}

	if err := c.sendFrame(headerData); err != nil {
		return err
	}

	if err := c.sendFrame(data); err != nil {
		return err
	}

	return nil
}

func (c *testClient) Receive() (ResponseHeader, []byte, error) {
	headerData, err := c.receiveFrame()
	if err != nil {
		return ResponseHeader{}, nil, err
	}

	var header ResponseHeader
	if err := json.Unmarshal(headerData, &header); err != nil {
		return ResponseHeader{}, nil, err
	}

	data, err := c.receiveFrame()
	if err != nil {
		return ResponseHeader{}, nil, err
	}

	return header, data, nil
}

func (c *testClient) sendFrame(frameData []byte) error {
	frameLen := uint32(len(frameData))
	if err := binary.Write(c.conn, binary.BigEndian, frameLen); err != nil {
		return err
	}

	if _, err := c.conn.Write(frameData); err != nil {
		return err
	}

	return nil
}

func (c *testClient) receiveFrame() ([]byte, error) {
	var frameLen uint32
	if err := binary.Read(c.conn, binary.BigEndian, &frameLen); err != nil {
		return nil, err
	}

	frameReader := io.LimitedReader{
		R: c.conn,
		N: int64(frameLen),
	}

	frameData, err := ioutil.ReadAll(&frameReader)
	if err != nil {
		return nil, err
	}

	return frameData, nil
}

func TestTransport(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	serverConn.SetDeadline(time.Now().Add(time.Second * 10))
	clientConn.SetDeadline(time.Now().Add(time.Second * 10))

	client := &testClient{conn: clientConn}
	transport := NewTransport(serverConn)

	sentHeader := RequestHeader{
		ApiVersion: "x.y.z",
		Endpoint:   "foo",
	}
	sentData := []byte(string("abcd1234"))
	go func() {
		err := client.Send(sentHeader, sentData)
		require.NoError(t, err)
	}()

	header, data, err := transport.Receive()
	require.NoError(t, err)
	assert.Equal(t, sentHeader, header)
	assert.Equal(t, sentData, data)

	replyHeader := ResponseHeader{Success: true}
	replyData := []byte(string("xyz56789"))
	go func() {
		err = transport.Respond(replyHeader, replyData)
		require.NoError(t, err)
	}()

	receivedHeader, receivedData, err := client.Receive()
	require.NoError(t, err)
	assert.Equal(t, replyHeader, receivedHeader)
	assert.Equal(t, replyData, receivedData)
}
