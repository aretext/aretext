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

func createListenSocket(socketPath string) (*net.UnixListener, error) {
	err := syscall.Unlink(socketPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("syscall.Unlink: %w", err)
	}

	addr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("net.ResolveUnixAddr: %w", err)
	}

	fmt.Printf("listening on %s\n", addr)
	ul, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, fmt.Errorf("net.ListenUnix: %w", err)
	}

	return ul, nil
}

func handleConnection(uc *net.UnixConn, clientId int) {
	fmt.Printf("client %d connected\n", clientId)

	// Receive client TTY file descriptor over the Unix socket.
	tty, err := receiveTTY(uc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error receiving tty: %s\n", err)
		return
	}
	defer tty.Close()

	fmt.Printf("received tty from client %d\n", clientId)

	// Start a shell subprocess connected to the client's tty.
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "/bin/bash", "--noprofile", "--norc")
	cmd.Env = []string{fmt.Sprintf("CLIENT_ID=%d", clientId)}
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty
	// https://github.com/golang/go/issues/29458
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
		Ctty:    0, // this must be a valid FD in the child process, so choose stdin (fd=0)
	}

	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running cmd: %s\n", err)
		return
	}

	fmt.Printf("cmd completed for client %d\n", clientId)
}

