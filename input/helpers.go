package input

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/input/engine"
)

func eventKeyToEngineEvent(eventKey *tcell.EventKey) engine.Event {
	if eventKey.Key() == tcell.KeyRune {
		return runeToEngineEvent(eventKey.Rune())
	} else {
		return keyToEngineEvent(eventKey.Key())
	}
}

func keyToEngineEvent(key tcell.Key) engine.Event {
	return engine.Event(int64(key) << 32)
}

func runeToEngineEvent(r rune) engine.Event {
	return engine.Event((int64(tcell.KeyRune) << 32) | int64(r))
}

func engineEventToKey(engineEvent engine.Event) tcell.Key {
	return tcell.Key(engineEvent >> 32)
}

func engineEventToRune(engineEvent engine.Event) rune {
	return rune(engineEvent & 0xFFFF)
}

const (
	captureIdVerbCount = engine.CaptureId(1<<16) + iota
	captureIdObjectCount
	captureIdClipboardPage
	captureIdMatchChar
	captureIdReplaceChar
	captureIdInsertChar
)

// Pre-compute and share these expressions to reduce number of allocations.
var verbCountExpr, objectCountExpr, clipboardPageExpr, matchCharExpr, replaceCharExpr, insertExpr engine.Expr

func init() {
	verbCountExpr = engine.OptionExpr{
		Child: engine.CaptureExpr{
			CaptureId: captureIdVerbCount,
			Child: engine.ConcatExpr{
				Children: []engine.Expr{
					engine.EventRangeExpr{
						StartEvent: runeToEngineEvent('1'),
						EndEvent:   runeToEngineEvent('9'),
					},
					engine.StarExpr{
						Child: engine.EventRangeExpr{
							StartEvent: runeToEngineEvent('0'),
							EndEvent:   runeToEngineEvent('9'),
						},
					},
				},
			},
		},
	}

	objectCountExpr = engine.OptionExpr{
		Child: engine.CaptureExpr{
			CaptureId: captureIdObjectCount,
			Child: engine.ConcatExpr{
				Children: []engine.Expr{
					engine.EventRangeExpr{
						StartEvent: runeToEngineEvent('1'),
						EndEvent:   runeToEngineEvent('9'),
					},
					engine.StarExpr{
						Child: engine.EventRangeExpr{
							StartEvent: runeToEngineEvent('0'),
							EndEvent:   runeToEngineEvent('9'),
						},
					},
				},
			},
		},
	}

	clipboardPageExpr = engine.OptionExpr{
		Child: engine.ConcatExpr{
			Children: []engine.Expr{
				engine.EventExpr{
					Event: runeToEngineEvent('"'),
				},
				engine.CaptureExpr{
					CaptureId: captureIdClipboardPage,
					Child: engine.EventRangeExpr{
						StartEvent: runeToEngineEvent('a'),
						EndEvent:   runeToEngineEvent('z'),
					},
				},
			},
		},
	}

	matchCharExpr = engine.CaptureExpr{
		CaptureId: captureIdMatchChar,
		Child: engine.EventRangeExpr{
			StartEvent: runeToEngineEvent(rune(0)),
			EndEvent:   runeToEngineEvent(rune(255)),
		},
	}

	replaceCharExpr = engine.CaptureExpr{
		CaptureId: captureIdReplaceChar,
		Child: engine.AltExpr{
			Children: []engine.Expr{
				engine.EventRangeExpr{
					StartEvent: runeToEngineEvent(rune(0)),
					EndEvent:   runeToEngineEvent(rune(255)),
				},
				engine.EventExpr{
					Event: keyToEngineEvent(tcell.KeyEnter),
				},
				engine.EventExpr{
					Event: keyToEngineEvent(tcell.KeyTab),
				},
			},
		},
	}

	insertExpr = engine.CaptureExpr{
		CaptureId: captureIdInsertChar,
		Child: engine.EventRangeExpr{
			StartEvent: runeToEngineEvent(rune(0)),
			EndEvent:   runeToEngineEvent(utf8.MaxRune),
		},
	}
}

type captureOpts struct {
	count         bool
	clipboardPage bool
	matchChar     bool
	replaceChar   bool
}

func altExpr(children ...engine.Expr) engine.Expr {
	return engine.AltExpr{Children: children}
}

func verbCountThenExpr(expr engine.Expr) engine.Expr {
	return engine.ConcatExpr{Children: []engine.Expr{verbCountExpr, expr}}
}

func runeExpr(r rune) engine.Expr {
	return engine.EventExpr{Event: runeToEngineEvent(r)}
}

func keyExpr(key tcell.Key) engine.Expr {
	return engine.EventExpr{Event: keyToEngineEvent(key)}
}

func cmdExpr(verb string, object string, opts captureOpts) engine.Expr {
	expr := engine.ConcatExpr{Children: make([]engine.Expr, 0, len(verb))}
	for _, r := range verb {
		expr.Children = append(expr.Children, engine.EventExpr{
			Event: runeToEngineEvent(r),
		})
	}

	if object != "" {
		verbExpr := expr
		objExpr := engine.ConcatExpr{Children: make([]engine.Expr, 0, len(object))}
		for _, r := range object {
			objExpr.Children = append(objExpr.Children, engine.EventExpr{
				Event: runeToEngineEvent(r),
			})
		}

		if opts.count {
			objExpr = engine.ConcatExpr{Children: []engine.Expr{objectCountExpr, objExpr}}
		}

		expr = engine.ConcatExpr{Children: []engine.Expr{verbExpr, objExpr}}
	}

	if opts.count {
		expr = engine.ConcatExpr{Children: []engine.Expr{verbCountExpr, expr}}
	}

	if opts.clipboardPage {
		expr = engine.ConcatExpr{Children: []engine.Expr{clipboardPageExpr, expr}}
	}

	if opts.matchChar {
		expr = engine.ConcatExpr{Children: []engine.Expr{expr, matchCharExpr}}
	}

	if opts.replaceChar {
		expr = engine.ConcatExpr{Children: []engine.Expr{expr, replaceCharExpr}}
	}

	return expr
}

func capturesToCommandParams(captures map[engine.CaptureId][]engine.Event) CommandParams {
	p := CommandParams{
		Count:         1,
		ClipboardPage: clipboard.PageDefault,
		MatchChar:     '\x00',
		ReplaceChar:   '\x00',
		InsertChar:    '\x00',
	}
	for captureId, captureEvents := range captures {
		switch captureId {
		case captureIdVerbCount, captureIdObjectCount:
			// Multiply here so that if both verb and object count are provided,
			// the total count is the product of the two counts.
			// For example, "2d3w" should delete 2*3=6 words.
			p.Count *= eventsToCount(captureEvents)
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

func eventsToCount(events []engine.Event) uint64 {
	var sb strings.Builder
	for _, e := range events {
		sb.WriteRune(engineEventToRune(e))
	}
	i, err := strconv.Atoi(sb.String())
	if err != nil || i < 0 {
		return 0
	}
	return uint64(i)
}

func eventsToClipboardPage(events []engine.Event) clipboard.PageId {
	if len(events) != 1 {
		return clipboard.PageNull
	}
	return clipboard.PageIdForLetter(engineEventToRune(events[0]))
}

func eventsToChar(events []engine.Event) rune {
	if len(events) != 1 {
		return '\x00'
	}
	return engineEventToRune(events[0])
}

func eventsToReplaceChar(events []engine.Event) rune {
	if len(events) != 1 {
		return '\x00'
	}

	switch engineEventToKey(events[0]) {
	case tcell.KeyEnter:
		return '\n'
	case tcell.KeyTab:
		return '\t'
	case tcell.KeyRune:
		return engineEventToRune(events[0])
	default:
		return '\x00'
	}
}
