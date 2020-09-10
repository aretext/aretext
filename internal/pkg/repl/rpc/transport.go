package rpc

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
)

// RequestHeader is metadata about a request to the server.
type RequestHeader struct {
	// ApiVersion is the version of the API used by the client.
	// The server will reject requests that don't match its API version.
	ApiVersion string `json:"api_version"`

	// ApiKey is a randomly generated key used to authorize clients.
	ApiKey string `json:"api_key"`

	// Endpoint is the endpoint targeted by the request.
	// The endpoint determines the format of the request and response messages.
	Endpoint string `json:"endpoint"`
}

// ResponseHeader is metadata about the response from the server.
type ResponseHeader struct {
	// Success indicates whether the RPC was completed successfully.
	Success bool `json:"success"`

	// Error describes an error that prevented the RPC from completing successfully.
	// If the RPC was successful, this will be an empty string.
	Error string `json:"error"`
}

// Transport communicates with a client.
type Transport struct {
	conn io.ReadWriteCloser
}

// NewTransport creates a new transport from the provided connection to the client.
// The transport will close the connection when the caller invokes Transport.Close().
func NewTransport(conn io.ReadWriteCloser) *Transport {
	return &Transport{conn}
}

// Receive receives a message from the client, blocking until a message is available.
func (t *Transport) Receive() (RequestHeader, []byte, error) {
	reqHeader, err := t.receiveHeader()
	if err != nil {
		return RequestHeader{}, nil, errors.Wrapf(err, "receiveHeader()")
	}

	reqData, err := t.receiveFrame()
	if err != nil {
		return RequestHeader{}, nil, errors.Wrapf(err, "receiveFrame()")
	}

	return reqHeader, reqData, nil
}

func (t *Transport) receiveHeader() (RequestHeader, error) {
	data, err := t.receiveFrame()
	if err != nil {
		return RequestHeader{}, errors.Wrapf(err, "receiveFrame()")
	}

	var header RequestHeader
	if err := json.Unmarshal(data, &header); err != nil {
		return RequestHeader{}, errors.Wrapf(err, "json.Unmarshal()")
	}

	return header, nil
}

func (t *Transport) receiveFrame() ([]byte, error) {
	var frameLen uint32
	if err := binary.Read(t.conn, binary.BigEndian, &frameLen); err != nil {
		return nil, errors.Wrapf(err, "binary.Read()")
	}

	frameReader := io.LimitedReader{
		R: t.conn,
		N: int64(frameLen),
	}

	frameData, err := ioutil.ReadAll(&frameReader)
	if err != nil {
		return nil, errors.Wrapf(err, "ioutil.ReadAll()")
	}

	return frameData, nil
}

// Respond sends a response message to the client.
func (t *Transport) Respond(header ResponseHeader, data []byte) error {
	if err := t.sendHeader(header); err != nil {
		return errors.Wrapf(err, "sendHeader()")
	}

	if header.Success {
		if err := t.sendFrame(data); err != nil {
			return errors.Wrapf(err, "sendFrame()")
		}
	}

	return nil
}

func (t *Transport) sendHeader(header ResponseHeader) error {
	headerData, err := json.Marshal(&header)
	if err != nil {
		return errors.Wrapf(err, "json.Marshal()")
	}
	return t.sendFrame(headerData)
}

func (t *Transport) sendFrame(frameData []byte) error {
	frameLen := uint32(len(frameData))
	if err := binary.Write(t.conn, binary.BigEndian, frameLen); err != nil {
		return errors.Wrapf(err, "binary.Write()")
	}

	if _, err := t.conn.Write(frameData); err != nil {
		return errors.Wrapf(err, "conn.Write()")
	}

	return nil
}

// Close terminates the connection to the client.
// Once the connection is closed, no further messages can be sent or received.
func (t *Transport) Close() {
	t.conn.Close()
}
