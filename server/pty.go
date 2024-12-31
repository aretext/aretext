package server

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

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
	// tcell won't exit its input loop until tty.Read returns.
	// To avoid waiting for input that will never arrive, set the pts to non-blocking.
	err := syscall.SetNonblock(int(tty.pts.Fd()), true)
	if err != nil {
		return fmt.Errorf("failed to set pts to non-blocking: %w", err)
	}

	// Platform-specific for Linux and BSD.
	err = drainPty(tty.pts)
	if err != nil {
		return fmt.Errorf("failed to drain pts: %w", err)
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
