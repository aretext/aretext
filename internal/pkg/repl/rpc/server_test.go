package rpc

import (
	"net"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeExecutor struct {
	apiVersion string
	respData   []byte
	execError  error
}

func (e *fakeExecutor) ApiVersion() string {
	return e.apiVersion
}

func (e *fakeExecutor) ExecuteAsync(endpoint string, data []byte) (chan []byte, error) {
	if e.execError != nil {
		return nil, e.execError
	}

	replyChan := make(chan []byte, 1)
	go func() { replyChan <- e.respData }()
	return replyChan, nil
}

func withClientAndServer(t *testing.T, executor AsyncExecutor, f func(*testing.T, *testClient, *Server)) {
	server, err := NewServer(executor)
	require.NoError(t, err)

	go server.ListenAndServe()
	defer server.Terminate()

	addr, err := server.Addr()
	require.NoError(t, err)

	clientConn, err := net.Dial(addr.Network(), addr.String())
	require.NoError(t, err)
	defer clientConn.Close()

	clientConn.SetDeadline(time.Now().Add(time.Second * 15))
	client := &testClient{conn: clientConn}

	f(t, client, server)
}

func TestServer(t *testing.T) {
	fakeApiVersion := "x.y.z"
	fakeRespData := []byte(string("xyz56789"))
	executor := &fakeExecutor{
		apiVersion: fakeApiVersion,
		respData:   fakeRespData,
	}

	withClientAndServer(t, executor, func(t *testing.T, client *testClient, server *Server) {
		reqHeader := RequestHeader{
			ApiVersion: fakeApiVersion,
			ApiKey:     server.ApiKey(),
			Endpoint:   "foobar",
		}
		reqData := []byte(string("abcd1234"))
		err := client.Send(reqHeader, reqData)
		require.NoError(t, err)

		respHeader, respData, err := client.Receive()
		require.NoError(t, err)
		assert.Equal(t, ResponseHeader{Success: true}, respHeader)
		assert.Equal(t, fakeRespData, respData)
	})
}

func TestServerRejectInvalidApiKey(t *testing.T) {
	executor := &fakeExecutor{
		apiVersion: "x.y.z",
		respData:   []byte{},
	}

	withClientAndServer(t, executor, func(t *testing.T, client *testClient, server *Server) {
		reqHeader := RequestHeader{
			ApiVersion: executor.apiVersion,
			ApiKey:     "invalid",
			Endpoint:   "foobar",
		}
		reqData := []byte(string("abcd1234"))
		err := client.Send(reqHeader, reqData)
		require.NoError(t, err)

		respHeader, _, err := client.Receive()
		require.NoError(t, err)
		expectedHeader := ResponseHeader{
			Success: false,
			Error:   "Unauthorized",
		}
		assert.Equal(t, expectedHeader, respHeader)
	})
}

func TestServerRejectInvalidApiVersion(t *testing.T) {
	executor := &fakeExecutor{
		apiVersion: "x.y.z",
		respData:   []byte{},
	}

	withClientAndServer(t, executor, func(t *testing.T, client *testClient, server *Server) {
		reqHeader := RequestHeader{
			ApiVersion: "invalid",
			ApiKey:     server.ApiKey(),
			Endpoint:   "foobar",
		}
		reqData := []byte(string("abcd1234"))
		err := client.Send(reqHeader, reqData)
		require.NoError(t, err)

		respHeader, _, err := client.Receive()
		require.NoError(t, err)
		expectedHeader := ResponseHeader{
			Success: false,
			Error:   "Unsupported API version",
		}
		assert.Equal(t, expectedHeader, respHeader)
	})
}

func TestServerExecutorError(t *testing.T) {
	executor := &fakeExecutor{
		apiVersion: "x.y.z",
		respData:   []byte{},
		execError:  errors.New("KABOOM!"),
	}

	withClientAndServer(t, executor, func(t *testing.T, client *testClient, server *Server) {
		reqHeader := RequestHeader{
			ApiVersion: executor.apiVersion,
			ApiKey:     server.ApiKey(),
			Endpoint:   "foobar",
		}
		reqData := []byte(string("abcd1234"))
		err := client.Send(reqHeader, reqData)
		require.NoError(t, err)

		respHeader, _, err := client.Receive()
		require.NoError(t, err)
		expectedHeader := ResponseHeader{
			Success: false,
			Error:   "KABOOM!",
		}
		assert.Equal(t, expectedHeader, respHeader)
	})
}
