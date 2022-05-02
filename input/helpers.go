package input

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/input/vm"
)

func eventKeyToVmEvent(eventKey *tcell.EventKey) vm.Event {
	if eventKey.Key() == tcell.KeyRune {
		return runeToVmEvent(eventKey.Rune())
	} else {
		return keyToVmEvent(eventKey.Key())
	}
}

func keyToVmEvent(key tcell.Key) vm.Event {
	return vm.Event(int64(key) << 32)
}

func runeToVmEvent(r rune) vm.Event {
	return vm.Event((int64(tcell.KeyRune) << 32) | int64(r))
}

func vmEventToKey(vmEvent vm.Event) tcell.Key {
	return tcell.Key(vmEvent >> 32)
}

func vmEventToRune(vmEvent vm.Event) rune {
	return rune(vmEvent & 0xFFFF)
}

const (
	captureIdCount = vm.CaptureId(1<<16) + iota
	captureIdClipboardPage
	captureIdMatchChar
	captureIdReplaceChar
	captureIdInsertChar
)

type captureOpts struct {
	count         bool
	clipboardPage bool
	matchChar     bool
	replaceChar   bool
}

func altExpr(children ...vm.Expr) vm.Expr {
	return vm.AltExpr{Children: children}
}

func runeExpr(r rune) vm.Expr {
	return vm.EventExpr{Event: runeToVmEvent(r)}
}

func keyExpr(key tcell.Key) vm.Expr {
	return vm.EventExpr{Event: keyToVmEvent(key)}
}

// Pre-compute and share these expressions to reduce number of allocations.
var countExpr, clipboardPageExpr, matchCharExpr, replaceCharExpr, insertExpr vm.Expr

func init() {
	countExpr = vm.OptionExpr{
		Child: vm.CaptureExpr{
			CaptureId: captureIdCount,
			Child: vm.ConcatExpr{
				Children: []vm.Expr{
					vm.EventRangeExpr{
						StartEvent: runeToVmEvent('1'),
						EndEvent:   runeToVmEvent('9'),
					},
					vm.StarExpr{
						Child: vm.EventRangeExpr{
							StartEvent: runeToVmEvent('0'),
							EndEvent:   runeToVmEvent('9'),
						},
					},
				},
			},
		},
	}

	clipboardPageExpr = vm.OptionExpr{
		Child: vm.ConcatExpr{
			Children: []vm.Expr{
				vm.EventExpr{
					Event: runeToVmEvent('"'),
				},
				vm.CaptureExpr{
					CaptureId: captureIdClipboardPage,
					Child: vm.EventRangeExpr{
						StartEvent: runeToVmEvent('a'),
						EndEvent:   runeToVmEvent('z'),
					},
				},
			},
		},
	}

	matchCharExpr = vm.CaptureExpr{
		CaptureId: captureIdMatchChar,
		Child: vm.EventRangeExpr{
			StartEvent: runeToVmEvent(rune(0)),
			EndEvent:   runeToVmEvent(rune(255)),
		},
	}

	replaceCharExpr = vm.CaptureExpr{
		CaptureId: captureIdReplaceChar,
		Child: vm.AltExpr{
			Children: []vm.Expr{
				vm.EventRangeExpr{
					StartEvent: runeToVmEvent(rune(0)),
					EndEvent:   runeToVmEvent(rune(255)),
				},
				vm.EventExpr{
					Event: keyToVmEvent(tcell.KeyEnter),
				},
				vm.EventExpr{
					Event: keyToVmEvent(tcell.KeyTab),
				},
			},
		},
	}

	insertExpr = vm.CaptureExpr{
		CaptureId: captureIdInsertChar,
		Child: vm.EventRangeExpr{
			StartEvent: runeToVmEvent(rune(0)),
			EndEvent:   runeToVmEvent(utf8.MaxRune),
		},
	}
}

func cmdExpr(verb string, object string, opts captureOpts) vm.Expr {
	expr := vm.ConcatExpr{Children: make([]vm.Expr, 0, len(verb))}
	for _, r := range verb {
		expr.Children = append(expr.Children, vm.EventExpr{
			Event: runeToVmEvent(r),
		})
	}

	if object != "" {
		verbExpr := expr
		objExpr := vm.ConcatExpr{Children: make([]vm.Expr, 0, len(object))}
		for _, r := range object {
			objExpr.Children = append(objExpr.Children, vm.EventExpr{
				Event: runeToVmEvent(r),
			})
		}
		expr = vm.ConcatExpr{Children: []vm.Expr{verbExpr, objExpr}}
	}

	if opts.count {
		expr = vm.ConcatExpr{Children: []vm.Expr{countExpr, expr}}
	}

	if opts.clipboardPage {
		expr = vm.ConcatExpr{Children: []vm.Expr{clipboardPageExpr, expr}}
	}

	if opts.matchChar {
		expr = vm.ConcatExpr{Children: []vm.Expr{expr, matchCharExpr}}
	}

	if opts.replaceChar {
		expr = vm.ConcatExpr{Children: []vm.Expr{expr, replaceCharExpr}}
	}

	return expr
}

func capturesToCommandParams(captures []vm.Capture, events []vm.Event) CommandParams {
	p := CommandParams{
		Count:         1,
		ClipboardPage: clipboard.PageDefault,
		MatchChar:     '\x00',
		ReplaceChar:   '\x00',
		InsertChar:    '\x00',
	}
	for _, capture := range captures {
		captureEvents := events[capture.StartIdx : capture.StartIdx+capture.Length]
		switch capture.Id {
		case captureIdCount:
			p.Count = eventsToCount(captureEvents)
		case captureIdClipboardPage:
			p.ClipboardPage = eventsToClipboardPage(captureEvents)
		case captureIdMatchChar:
			p.MatchChar = eventsToChar(captureEvents)
		case captureIdReplaceChar:
			p.ReplaceChar = eventsToReplaceChar(captureEvents)
		case captureIdInsertChar:
			p.InsertChar = eventsToChar(captureEvents)
		}
	}
	return p
}

func eventsToCount(events []vm.Event) uint64 {
	var sb strings.Builder
	for _, e := range events {
		sb.WriteRune(vmEventToRune(e))
	}
	i, err := strconv.Atoi(sb.String())
	if err != nil || i < 0 {
		return 0
	}
	return uint64(i)
}

func eventsToClipboardPage(events []vm.Event) clipboard.PageId {
	if len(events) != 1 {
		return clipboard.PageNull
	}
	return clipboard.PageIdForLetter(vmEventToRune(events[0]))
}

func eventsToChar(events []vm.Event) rune {
	if len(events) != 1 {
		return '\x00'
	}
	return vmEventToRune(events[0])
}

func eventsToReplaceChar(events []vm.Event) rune {
	if len(events) != 1 {
		return '\x00'
	}

	switch vmEventToKey(events[0]) {
	case tcell.KeyEnter:
		return '\n'
	case tcell.KeyTab:
		return '\t'
	case tcell.KeyRune:
		return vmEventToRune(events[0])
	default:
		return '\x00'
	}
}
