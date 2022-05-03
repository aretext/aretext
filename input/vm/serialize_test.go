package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgramSerialization(t *testing.T) {
	prog := MustCompile(ConcatExpr{
		Children: []Expr{
			EventExpr{Event: 123},
			EventRangeExpr{
				StartEvent: -456,
				EndEvent:   78910,
			},
		},
	})

	data := SerializeProgram(prog)
	result := DeserializeProgram(data)
	assert.Equal(t, prog, result)
}
