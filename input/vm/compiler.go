package vm

import "fmt"

// op represents a bytecode for the input virtual machine.
type op uint8

const (
	// opNone is a placeholder bytecode.
	// It should never appear in the output of a compiled program,
	// and the runtime will panic if it sees this.
	opNone = op(iota)

	// opRead reads the next input event in a range.
	//  * arg1 represents the start of the range (inclusive).
	//  * arg2 represents the end of the range (inclusive).
	//
	// If the next input event is within the range, the thread
	// continues from the next instruction in the program.
	// If the next input event is NOT within the range, the thread
	// terminates without accepting the input.
	// If there are no more events to read, the thread blocks waiting
	// for the next input event.
	opRead

	// opJump jumps to a different position in the program.
	//  * arg1 represents the program counter (int)
	//  * arg2 is not used.
	opJump

	// opFork splits the current thread into two threads at
	// different positions in the program.
	//  * arg1 represents the program counter for the first thread (int)
	//  * arg2 represents the program counter for the second thread (int)
	//
	// The runtime executes the first thread before the second.
	opFork

	// opStartCapture begins capturing input from the current event onward.
	//  * arg1 represents the capture ID (int)
	//  * arg2 is not used.
	opStartCapture

	// opEndCapture completes a capture at the current event.
	// The capture ID MUST match a previously started capture,
	// or the runtime will panic.
	//  * arg1 represents the capture ID (int)
	//  * arg2 is not used.
	opEndCapture

	// opAccept accepts the input.
	// This should always appear as the last instruction in the program and nowhere else.
	// Neither arg1 nor arg2 are used.
	opAccept
)

// bytecode is a single instruction in a program.
type bytecode struct {
	op   op
	arg1 int64
	arg2 int64
}

// Program controls how the virtual machine runtime processes input events.
// Valid programs are capable of recognizing any regular language.
type Program []bytecode

// MustCompile panics if compilation fails.
func MustCompile(expr Expr) Program {
	program, err := Compile(expr)
	if err != nil {
		panic(err)
	}
	return program
}

// Compile transforms a regular expression to a program executable by the virtual machine runtime.
func Compile(expr Expr) (Program, error) {
	if err := validateExpr(expr); err != nil {
		return nil, err
	}

	prog := compileRecursively(expr, nil)
	prog = append(prog, bytecode{op: opAccept}) // last instruction is always opAccept.
	return prog, nil
}

func validateExpr(expr Expr) error {
	var validateRecursively func(Expr, []CaptureId) error
	validateRecursively = func(expr Expr, parentCaptureIds []CaptureId) error {
		switch expr := expr.(type) {
		case EventExpr:
			break
		case EventRangeExpr:
			if expr.StartEvent >= expr.EndEvent {
				return fmt.Errorf("Invalid event range [%d, %d]", expr.StartEvent, expr.EndEvent)
			}
		case ConcatExpr:
			for _, child := range expr.Children {
				if err := validateRecursively(child, parentCaptureIds); err != nil {
					return err
				}
			}
		case AltExpr:
			for _, child := range expr.Children {
				if err := validateRecursively(child, parentCaptureIds); err != nil {
					return err
				}
			}
		case OptionExpr:
			return validateRecursively(expr.Child, parentCaptureIds)
		case StarExpr:
			return validateRecursively(expr.Child, parentCaptureIds)
		case CaptureExpr:
			for _, id := range parentCaptureIds {
				if id == expr.CaptureId {
					return fmt.Errorf("Conflicting capture ID %d", expr.CaptureId)
				}
			}
			return validateRecursively(expr.Child, append(parentCaptureIds, expr.CaptureId))
		default:
			return fmt.Errorf("Invalid expression type %T", expr)
		}

		return nil
	}

	var captureIds []CaptureId
	return validateRecursively(expr, captureIds)
}

func compileRecursively(expr Expr, prog Program) Program {
	switch expr := expr.(type) {
	case EventExpr:
		return append(prog, bytecode{
			op:   opRead,
			arg1: int64(expr.Event),
			arg2: int64(expr.Event),
		})

	case EventRangeExpr:
		return append(prog, bytecode{
			op:   opRead,
			arg1: int64(expr.StartEvent),
			arg2: int64(expr.EndEvent),
		})

	case ConcatExpr:
		for _, child := range expr.Children {
			prog = compileRecursively(child, prog)
		}
		return prog

	case AltExpr:
		if len(expr.Children) == 0 {
			return prog
		} else if len(expr.Children) == 1 {
			return compileRecursively(expr.Children[0], prog)
		} else if len(expr.Children) > 1 {
			// Split off the first child. The remaining children
			// will be compiled recursively as an AltExpr.
			leftChild := expr.Children[0]
			rightChildren := expr.Children[1:len(expr.Children)]

			// Placeholder for opFork, which will be filled in later.
			forkIdx := len(prog)
			prog = append(prog, bytecode{})

			// Compile left child recursively.
			leftChildIdx := len(prog)
			prog = compileRecursively(leftChild, prog)

			// Placeholder for opJump, which will be filled in later.
			jumpIdx := len(prog)
			prog = append(prog, bytecode{})

			// Compile right children recursively as an AltExpr.
			rightChildIdx := len(prog)
			rightExpr := AltExpr{Children: rightChildren}
			prog = compileRecursively(rightExpr, prog)
			endIdx := len(prog)

			// Fill in opFork now that we know the program counters for the left and right child.
			prog[forkIdx] = bytecode{
				op:   opFork,
				arg1: int64(leftChildIdx),
				arg2: int64(rightChildIdx),
			}

			// Fill in opJump now that we know the length of this part of the program.
			prog[jumpIdx] = bytecode{
				op:   opJump,
				arg1: int64(endIdx),
			}
		}

	case OptionExpr:
		// Placeholder for opFork, which will be filled in later.
		forkIdx := len(prog)
		prog = append(prog, bytecode{})

		// Compile the child recursively.
		childStartIdx := len(prog)
		prog = compileRecursively(expr.Child, prog)
		endIdx := len(prog)

		// Fill in opFork now that we know the length of this part of the program.
		prog[forkIdx] = bytecode{
			op:   opFork,
			arg1: int64(childStartIdx),
			arg2: int64(endIdx),
		}

	case StarExpr:
		// Placeholder for opFork, which will be filled in later.
		forkIdx := len(prog)
		prog = append(prog, bytecode{})

		// Compile the child recursively.
		childStartIdx := len(prog)
		prog = compileRecursively(expr.Child, prog)

		// Jump back to the first instruction so repetitions are accepted.
		prog = append(prog, bytecode{
			op:   opJump,
			arg1: int64(forkIdx),
		})

		// Fill in opFork now that we know the length of this part of the program.
		endIdx := len(prog)
		prog[forkIdx] = bytecode{
			op:   opFork,
			arg1: int64(childStartIdx),
			arg2: int64(endIdx),
		}

	case CaptureExpr:
		// Start the capture.
		prog = append(prog, bytecode{
			op:   opStartCapture,
			arg1: int64(expr.CaptureId),
		})

		// Compile the child recursively.
		prog = compileRecursively(expr.Child, prog)

		// End the capture after the child completes.
		prog = append(prog, bytecode{
			op:   opEndCapture,
			arg1: int64(expr.CaptureId),
		})

	default:
		panic(fmt.Sprintf("Invalid expression type %T", expr))
	}

	return prog
}
