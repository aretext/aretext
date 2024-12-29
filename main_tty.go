// TODO: this is for testing, remove once verified.
package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"

	"github.com/aretext/aretext/pty"
)

func main() {
	ptmx, pts, err := pty.CreatePtyPair()
	if err != nil {
		panic(err)
	}

	tcellTty, err := pty.NewTtyFromPts(pts)
	if err != nil {
		panic(err)
	}
	defer tcellTty.Close()

	s, err := tcell.NewTerminfoScreenFromTtyTerminfo(tcellTty, nil)
	if err != nil {
		panic(err)
	}

	err = s.Init()
	if err != nil {
		panic(err)
	}
	defer s.Fini()

	defStyle := tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)
	s.SetStyle(defStyle)

	displayHelloWorld(s)

	quitChan := make(chan struct{})
	go runTcellEventLoop(s, quitChan)
	pty.ProxyTtyToPtmxUntilClosed(ptmx)

	select {
		case <-quitChan:
		// wait until tcell finishes cleanup
	}
}

func displayHelloWorld(s tcell.Screen) {
	w, h := s.Size()
	s.Clear()
	style := tcell.StyleDefault.Foreground(tcell.ColorCadetBlue.TrueColor()).Background(tcell.ColorWhite)
	emitStr(s, w/2-7, h/2, style, "Hello, World!")
	emitStr(s, w/2-9, h/2+1, tcell.StyleDefault, "Press ESC to exit.")
	s.Show()
}

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
}

func runTcellEventLoop(s tcell.Screen, quitChan chan struct{}) {
	for {
		switch ev := s.PollEvent().(type) {
		case *tcell.EventResize:
			s.Sync()
			displayHelloWorld(s)
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape {
				s.Fini()
				quitChan <- struct{}{}
				return
			}
		}
	}
}
