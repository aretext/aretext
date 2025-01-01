package clientserver

import (
	"errors"
	"fmt"
	"os"
	"sync"
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
	_ = drainTty(int(tty.pts.Fd())) // macOS loses output without this.
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
	// tcell won't exit its input loop until tty.Read returns.
	// To avoid waiting for input that will never arrive, set the pts to non-blocking.
	return setTtyNonblockAndDrain(int(tty.pts.Fd()))
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
