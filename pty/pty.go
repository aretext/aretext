package pty

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func CreatePtyPair() (ptmx *os.File, pts *os.File, err error) {
	// Create the pty pair.
	ptmxFd, err := unix.Open("/dev/ptmx", os.O_RDWR, 0o600)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open /dev/ptmx: %w", err)
	}

	// Unlock pts.
	locked := 0
	result, _, err := unix.Syscall(unix.SYS_IOCTL, uintptr(ptmxFd), unix.TIOCSPTLCK, uintptr(unsafe.Pointer(&locked)))
	if int(result) == -1 {
		return nil, nil, fmt.Errorf("could not unlock pty: %w", err)
	}

	// Retrieve pts file descriptor.
	ptsFd, _, err := unix.Syscall(unix.SYS_IOCTL, uintptr(ptmxFd), unix.TIOCGPTPEER, unix.O_RDWR|unix.O_NOCTTY)
	if int(ptsFd) == -1 {
		if errno, isErrno := err.(syscall.Errno); !isErrno || (errno != syscall.EINVAL && errno != syscall.ENOTTY) {
			return nil, nil, fmt.Errorf("could not retrieve pts file descriptor: %w", err)
		}
		// On EINVAL or ENOTTY, fallback to TIOCGPTN.
		ptyN, err := unix.IoctlGetInt(ptmxFd, unix.TIOCGPTN)
		if err != nil {
			return nil, nil, fmt.Errorf("could not find pty number: %w", err)
		}
		ptyName := fmt.Sprintf("/dev/pts/%d", ptyN)
		fd, err := unix.Open(ptyName, unix.O_RDWR|unix.O_NOCTTY, 0o620)
		if err != nil {
			return nil, nil, fmt.Errorf("could not open pty %s: %w", ptyName, err)
		}
		ptsFd = uintptr(fd)
	}

	ptmx = os.NewFile(uintptr(ptmxFd), "")
	pts = os.NewFile(ptsFd, "")
	return ptmx, pts, nil
}

func ResizePtyToMatchTty(tty *os.File, ptmx *os.File) (width, height int, err error) {
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

func ProxyTtyToPtmxUntilClosed(ptmx *os.File) {
	doneCh := make(chan struct{})

	// Copy tty input -> pty
	go func(ptyOut io.Writer, ttyIn io.Reader) {
		_, _ = io.Copy(ptyOut, ttyIn)
		doneCh <- struct{}{}
	}(ptmx, os.Stdin)

	// Copy pty output -> tty
	go func(ttyOut io.Writer, ptyIn io.Reader) {
		_, _ = io.Copy(ttyOut, ptyIn)
		doneCh <- struct{}{}
	}(os.Stdout, ptmx)

	// Block until pty closed (either by client or server).
	select {
	case <-doneCh:
	}

	return nil
}

// TODO
type ptsTty struct {
	pts   *os.File
	saved *term.State
	stopQ chan struct{}
	wg    sync.WaitGroup
	l     sync.Mutex
}

// TODO: explain this
func NewTtyFromPts(pts *os.File) (tcell.Tty, error) {
	var err error
	tty := &ptsTty{pts: pts}
	if !term.IsTerminal(int(pts.Fd())) {
		return nil, errors.New("not a terminal")
	}
	if tty.saved, err = term.GetState(int(pts.Fd())); err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}
	return tty, nil
}

func (tty *ptsTty) Read(b []byte) (int, error) {
	return tty.pts.Read(b)
}

func (tty *ptsTty) Write(b []byte) (int, error) {
	return tty.pts.Write(b)
}

func (tty *ptsTty) Close() error {
	return tty.pts.Close()
}

func (tty *ptsTty) Start() error {
	tty.l.Lock()
	defer tty.l.Unlock()

	var err error
	_ = tty.pts.SetReadDeadline(time.Time{})
	saved, err := term.MakeRaw(int(tty.pts.Fd())) // also sets vMin and vTime
	if err != nil {
		return err
	}
	tty.saved = saved

	tty.stopQ = make(chan struct{})
	tty.wg.Add(1)
	go func(stopQ chan struct{}) {
		defer tty.wg.Done()
		for {
			select {
			case <-stopQ:
				return
			}
		}
	}(tty.stopQ)

	return nil
}

func (tty *ptsTty) Drain() error {
	_ = tty.pts.SetReadDeadline(time.Now())

	_ = syscall.SetNonblock(int(tty.pts.Fd()), true)
	tio, err := unix.IoctlGetTermios(int(tty.pts.Fd()), unix.TCGETS)
	if err != nil {
		return err
	}
	tio.Cc[unix.VMIN] = 0
	tio.Cc[unix.VTIME] = 0
	if err = unix.IoctlSetTermios(int(tty.pts.Fd()), unix.TCSETSW, tio); err != nil {
		return err
	}

	return nil
}

func (tty *ptsTty) Stop() error {
	tty.l.Lock()
	if err := term.Restore(int(tty.pts.Fd()), tty.saved); err != nil {
		tty.l.Unlock()
		return err
	}
	_ = tty.pts.SetReadDeadline(time.Now())

	close(tty.stopQ)
	tty.l.Unlock()

	tty.wg.Wait()

	return nil
}

func (tty *ptsTty) WindowSize() (tcell.WindowSize, error) {
	ws, err := unix.IoctlGetWinsize(int(tty.pts.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return tcell.WindowSize{}, err
	}
	size := tcell.WindowSize{
		Width:       int(ws.Col),
		Height:      int(ws.Row),
		PixelWidth:  int(ws.Xpixel),
		PixelHeight: int(ws.Ypixel),
	}
	return size, nil
}

func (tty *ptsTty) NotifyResize(_ func()) {
	// Not implemented as pts won't receive SIGWINCH
}

var _ tcell.Tty = (*ptsTty)(nil)
