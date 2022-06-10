package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyProgram(t *testing.T) {
	testCases := []struct {
		name   string
		prog   Program
		errMsg string
	}{
		{
			name: "single accept op",
			prog: []bytecode{
				{op: opAccept},
			},
		},
		{
			name: "loop with read",
			prog: []bytecode{
				{op: opFork, arg1: 1, arg2: 3},
				{op: opRead},
				{op: opJump, arg1: 0},
				{op: opAccept},
			},
		},
		{
			name: "infinite loop",
			prog: []bytecode{
				{op: opJump, arg1: 0},
			},
			errMsg: "program loop must contain at least one read: [0]",
		},
		{
			name: "invalid bytecode op",
			prog: []bytecode{
				{op: op(14)},
			},
			errMsg: "invalid bytecode op 14",
		},
		{
			name: "bytecode opNone",
			prog: []bytecode{
				{op: opNone},
			},
			errMsg: "bytecode opNone is not allowed",
		},
		{
			name: "negative jump target",
			prog: []bytecode{
				{op: opJump, arg1: -1},
				{op: opAccept},
			},
			errMsg: "program target -1 is negative",
		},
		{
			name: "jump target past end of program",
			prog: []bytecode{
				{op: opJump, arg1: 3},
				{op: opAccept},
			},
			errMsg: "program target 3 is past end of program",
		},
		{
			name: "negative fork target",
			prog: []bytecode{
				{op: opFork, arg1: -1, arg2: 1},
				{op: opAccept},
			},
			errMsg: "program target -1 is negative",
		},
		{
			name: "fork target past end of program",
			prog: []bytecode{
				{op: opFork, arg1: 1, arg2: 2},
				{op: opAccept},
			},
			errMsg: "program target 2 is past end of program",
		},
		{
			name: "unreachable bytecode",
			prog: []bytecode{
				{op: opRead},
				{op: opRead},
				{op: opJump, arg1: 0},
				{op: opAccept},
			},
			errMsg: "program bytecode 3 is not reachable",
		},
		{
			name: "no accept at end",
			prog: []bytecode{
				{op: opRead},
				{op: opRead},
				{op: opRead},
			},
			errMsg: "program target 3 is past end of program",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := VerifyProgram(tc.prog)
			if tc.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errMsg)
			}
		})
	}
}
