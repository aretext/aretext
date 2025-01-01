package clientserver

import (
	"fmt"
	"log"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func setTtyRaw(tty *os.File) (restoreTty func(), err error) {
	// Set tty to raw mode.
	ttyFd := int(tty.Fd())
	oldTtyState, err := term.MakeRaw(ttyFd)
	if err != nil {
		return nil, fmt.Errorf("failed to set tty state: %w", err)
	}

	// Restore original state.
	restoreTty = func() {
		term.Restore(ttyFd, oldTtyState)
	}

	return restoreTty, nil
}

func getTtySize(tty *os.File) (width, height int, err error) {
	ws, err := unix.IoctlGetWinsize(int(tty.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, fmt.Errorf("unix.IoctlGetWinsize: %w", err)
	}
	return int(ws.Col), int(ws.Row), nil
}

// TODO
type RemoteTty struct {
	mu            sync.Mutex
	f             *os.File // unix domain socket
	width, height int
	cb            func()
}

// TODO: explain this
func NewRemoteTty(f *os.File, width int, height int) (*RemoteTty, error) {
	// TODO explain this: https://go-review.googlesource.com/c/go/+/81636
	fd := int(f.Fd())
	fd2, err := syscall.Dup(fd)
	if err != nil {
		return nil, fmt.Errorf("syscall.Dup: %w", err)
	}
	err = syscall.SetNonblock(fd2, true)
	if err != nil {
		return nil, fmt.Errorf("syscall.SetNonblock: %w", err)
	}

	f2 := os.NewFile(uintptr(fd2), "")

	return &RemoteTty{
		f:      f2,
		width:  width,
		height: height,
	}, nil
}

func (tty *RemoteTty) Read(b []byte) (int, error) {
	return tty.f.Read(b)
}

func (tty *RemoteTty) Write(b []byte) (int, error) {
	return tty.f.Write(b)
}

func (tty *RemoteTty) Close() error {
	return tty.f.Close()
}

func (tty *RemoteTty) Start() error {
	// in case we set nonblocking in Drain().
	var t time.Time
	err := tty.f.SetReadDeadline(t)
	if err != nil {
		return fmt.Errorf("file.SetReadDeadline: %w", err)
	}
	return nil
}

func (tty *RemoteTty) Drain() error {
	log.Printf("draining remote tty\n")

	// Set a read deadline otherwise tcell will deadlock waiting on input.
	err := tty.f.SetReadDeadline(time.Now())
	if err != nil {
		log.Printf("could not set read deadline: %s\n", err)
		return fmt.Errorf("file.SetReadDeadline: %w", err)
	}

	log.Printf("finished draining remote tty\n")
	return nil
}

func (tty *RemoteTty) Stop() error {
	// Nothing else to do here, as client tty can stay in raw mode
	// and we use a pty for subcommand.
	return nil
}

func (tty *RemoteTty) WindowSize() (tcell.WindowSize, error) {
	tty.mu.Lock()
	ws := tcell.WindowSize{
		Width:  tty.width,
		Height: tty.height,
	}
	tty.mu.Unlock()
	return ws, nil
}

func (tty *RemoteTty) Resize(width, height int) {
	tty.mu.Lock()
	tty.width = width
	tty.height = height
	tty.mu.Unlock()

	if tty.cb != nil {
		tty.cb()
	}
}

func (tty *RemoteTty) NotifyResize(cb func()) {
	tty.mu.Lock()
	tty.cb = cb
	tty.mu.Unlock()
}
