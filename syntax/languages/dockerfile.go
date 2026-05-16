package languages

import (
	"github.com/aretext/aretext/syntax/parser"
)

func DockerfileParseFunc() parser.Func {
	// TODO
	return func(parser.TrackingRuneIter, parser.State) parser.Result {
		return parser.Result{}
	}
}
