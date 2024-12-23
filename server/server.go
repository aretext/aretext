package server

// RunServer starts an aretext server.
// The server listens on a Unix Domain Socket (UDS) for clients to connect.
// The client sends the server a pseudoterminal (pty), which the server uses
// for input/output from/to the client's terminal.
func RunServer(config Config) error {
	releaseLock, err := acquireLock(config.LockPath)
	if err != nil {
		return fmt.Errorf("acquireLock: %w", err)
	}
	defer releaseLock()

	ul, err := createListenSocket(config.SocketPath)
	if err != nil {
		return fmt.Errorf("createListenSocket: %w", err)
	}

	clientId := 0
	for {
		conn, err := ul.AcceptUnix()
		if err != nil {
			return err
		}

		go handleConnection(conn, clientId)
		clientId++
	}
}
