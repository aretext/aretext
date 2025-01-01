package clientserver

import (
	"fmt"
	"os"
	"sync"

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

func resizePtyToMatchTty(tty *os.File, ptmx *os.File) (width, height int, err error) {
	// Update ptmx with the same size as client tty.
	ws, err := unix.IoctlGetWinsize(int(tty.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, fmt.Errorf("unix.IoctlGetWinsize: %w", err)
	}

	err = unix.IoctlSetWinsize(int(ptmx.Fd()), unix.TIOCSWINSZ, ws)
	if err != nil {
		return 0, 0, fmt.Errorf("unix.IoctlSetWinsize: %w", err)
	}

	return int(ws.Col), int(ws.Row), nil
}

func createPtyPair() (ptmx *os.File, pts *os.File, err error) {
	// Create the pty pair.
	ptmxFd, err := unix.Open("/dev/ptmx", os.O_RDWR, 0o600)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open /dev/ptmx: %w", err)
	}

	// Unlock pts.
	err = unlockPts(ptmxFd)
	if err != nil {
		return nil, nil, err
	}

	// File descriptors for both ptmx and pts.
	ptmx = os.NewFile(uintptr(ptmxFd), "")
	pts, err = ptsFileFromPtmx(ptmx)
	if err != nil {
		return nil, nil, err
	}
	return ptmx, pts, nil
}

// TODO
type RemoteTty struct {
	mu              sync.Mutex
	pipeIn, pipeOut *os.File
	width, height   int
}

// TODO: explain this
func NewRemoteTty(pipeIn *os.File, pipeOut *os.File, width int, height int) *RemoteTty {
	return &RemoteTty{
		pipeIn:  pipeIn,
		pipeOut: pipeOut,
		width:   width,
		height:  height,
	}
}

func (tty *RemoteTty) Read(b []byte) (int, error) {
	return tty.pipeIn.Read(b)
}

func (tty *RemoteTty) Write(b []byte) (int, error) {
	return tty.pipeOut.Write(b)
}

func (tty *RemoteTty) Close() error {
	// no-op since we don't own the pipe FD.
	return nil
}

func (tty *RemoteTty) Start() error {
	// TODO: how to restore original terminal settings?
	return nil
}

func (tty *RemoteTty) Drain() error {
	return nil
}

func (tty *RemoteTty) Stop() error {
	// TODO: how to restore original terminal settings?
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
}

func (tty *RemoteTty) NotifyResize(_ func()) {
	// Not implemented as pts won't receive SIGWINCH
}
