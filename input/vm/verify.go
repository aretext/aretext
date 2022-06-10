package vm

import "fmt"

// VerifyProgram checks that a program is valid.
// This is mainly used in tests to verify that serialized/embedded
// programs are safe to execute.
func VerifyProgram(prog Program) error {
	if len(prog) == 0 {
		return fmt.Errorf("program must have at least one bytecode")
	}

	edges := buildExecutionGraph(prog)
	seen := make(map[int]struct{}, len(prog))
	var verifyRecursively func(int, []int) error
	verifyRecursively = func(i int, path []int) error {
		if i < 0 {
			return fmt.Errorf("program target %d is negative", i)
		}

		if i >= len(prog) {
			return fmt.Errorf("program target %d is past end of program", i)
		}

		var isLoop, hasRead bool
		for _, j := range path {
			hasRead = hasRead || (prog[j].op == opRead)
			isLoop = isLoop || (j == i)
		}

		if isLoop && !hasRead {
			return fmt.Errorf("program loop must contain at least one read: %v", path)
		}

		if _, ok := seen[i]; ok {
			return nil
		}
		seen[i] = struct{}{}

		if err := verifyBytecode(prog[i]); err != nil {
			return err
		}

		for _, j := range edges[i] {
			if err := verifyRecursively(j, append(path, i)); err != nil {
				return err
			}
		}

		return nil
	}

	if err := verifyRecursively(0, nil); err != nil {
		return err
	}

	for i := 0; i < len(prog); i++ {
		if _, ok := seen[i]; !ok {
			return fmt.Errorf("program bytecode %d is not reachable", i)
		}
	}

	return nil
}

func buildExecutionGraph(prog Program) [][]int {
	edges := make([][]int, len(prog))
	for i, bc := range prog {
		switch bc.op {
		case opRead, opStartCapture, opEndCapture:
			edges[i] = []int{i + 1}
		case opJump:
			edges[i] = []int{int(bc.arg1)}
		case opFork:
			edges[i] = []int{int(bc.arg1), int(bc.arg2)}
		}
	}
	return edges
}

func verifyBytecode(bc bytecode) error {
	switch bc.op {
	case opNone:
		return fmt.Errorf("bytecode opNone is not allowed")
	case opRead, opJump, opFork, opStartCapture, opEndCapture, opAccept:
		// Allowed bytecode ops.
		return nil
	default:
		return fmt.Errorf("invalid bytecode op %d", bc.op)
	}
}
