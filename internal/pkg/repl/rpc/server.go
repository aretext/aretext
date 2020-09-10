package rpc

//go:generate go run ./gen.go

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net"

	"github.com/pkg/errors"
)

// Server schedules tasks based on remote procedure calls.
type Server struct {
	listener net.Listener
	executor AsyncExecutor
	apiKey   string
}

// NewServer creates a new server with the specified scheduler implementation.
func NewServer(executor AsyncExecutor) (*Server, error) {
	// Since we don't specify a port, the OS will randomly assign one that's available.
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return nil, errors.Wrapf(err, "net.Listen()")
	}

	// Randomly generate an API key.
	apiKey, err := generateApiKey()
	if err != nil {
		return nil, errors.Wrapf(err, "generateApiKey()")
	}

	server := &Server{
		listener: listener,
		executor: executor,
		apiKey:   apiKey,
	}

	return server, nil
}

// ListenAndServe listens on a localhost port for RPCs.
// The port is randomly assigned and can be retrieved using Addr().
// RPCs are passed to the task scheduler to schedule and execute tasks.
// This blocks until the server has been terminated.
func (s *Server) ListenAndServe() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return errors.Wrapf(err, "listener.Accept()")
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	transport := NewTransport(conn)
	defer transport.Close()
	for {
		requestHeader, requestData, err := transport.Receive()
		if err != nil {
			log.Printf("Error receiving from client: %v\n", err)
			return
		}

		if s.apiKey != requestHeader.ApiKey {
			apiKeyError := errors.New("Unauthorized")
			if err := s.sendErrorResponse(transport, apiKeyError); err != nil {
				log.Printf("Error sending response to client: %v\n", err)
				return
			}
		}

		if s.executor.ApiVersion() != requestHeader.ApiVersion {
			apiVersionError := errors.New("Unsupported API version")
			if err := s.sendErrorResponse(transport, apiVersionError); err != nil {
				log.Printf("Error sending response to client: %v\n", err)
				return
			}
		}

		responseChan, execError := s.executor.ExecuteAsync(requestHeader.Endpoint, requestData)
		if execError != nil {
			if err := s.sendErrorResponse(transport, execError); err != nil {
				log.Printf("Error sending response to client: %v\n", err)
				return
			}
		}

		responseData := <-responseChan
		responseHeader := ResponseHeader{Success: true}
		if err := transport.Respond(responseHeader, responseData); err != nil {
			log.Printf("Error sending response to client: %v\n", err)
			return
		}
	}
}

func (s *Server) sendErrorResponse(transport *Transport, err error) error {
	responseHeader := ResponseHeader{
		Success: false,
		Error:   err.Error(),
	}
	return transport.Respond(responseHeader, nil)
}

// Addr returns the network address the server is listening on.
func (s *Server) Addr() (net.Addr, error) {
	return s.listener.Addr(), nil
}

// ApiKey returns a randomly generated key clients must send to access the API.
func (s *Server) ApiKey() string {
	return s.apiKey
}

// Terminate tells the server to stop listening for requests.
func (s *Server) Terminate() error {
	return s.listener.Close()
}

func generateApiKey() (string, error) {
	randomBytes := make([]byte, 64)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", errors.Wrapf(err, "rand.Read()")
	}

	key := base64.StdEncoding.EncodeToString(randomBytes)
	return key, nil
}
