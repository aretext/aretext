package vm

// Event is an input event.
// This usually represents a keypress, but the input VM doesn't assume
// that the events have any particular meaning.
type Event int64

// CaptureId is an identifier for a subsequence of events.
type CaptureId int

// Capture is a subsequence of events from an accepted sequence.
// These are specified in the regular expression used to compile the VM program,
// but the input VM doesn't assume they have any particular meaning.
type Capture struct {
	Id       CaptureId
	StartIdx int // Relative to when the runtime was created or last reset.
	Length   int
}

// Result is the result of processing a single input event.
type Result struct {
	Accepted bool      // The input sequence was accepted.
	Reset    bool      // The VM reset to the beginning of the program (always true if Accepted is true).
	Captures []Capture // Subsequences captured for an accepted input sequence.
}

// Runtime is a virtual machine that runs a program to interpret input events.
// It is capable of recognizing any regular language.
// The implementation is heavily inspired by Russ Cox's
// "Regular Expression Matching: the Virtual Machine Approach"
// https://swtch.com/~rsc/regexp/regexp2.html
type Runtime struct {
	program        Program
	eventCount     int
	activeThreads  []threadState
	blockedThreads []threadState
}

func NewRuntime(program Program) *Runtime {
	initialThread := threadState{}
	return &Runtime{
		program:        program,
		eventCount:     0,
		activeThreads:  nil,
		blockedThreads: []threadState{initialThread},
	}
}

// ProcessEvent processes a single input event according to the runtime's program.
// If any thread accepts the input, the runtime accepts the input and resets.
// If all threads reject the input, the runtime resets.
// Otherwise, the runtime waits for more input before making a decision.
func (r *Runtime) ProcessEvent(event Event) Result {
	// Blocked threads become unblocked now that we have a new event.
	r.activeThreads, r.blockedThreads = r.blockedThreads, r.activeThreads
	r.blockedThreads = r.blockedThreads[:0]

	// Count this event so we can figure out the capture start/end indices.
	r.eventCount++

	// Run each thread until it either completes or blocks.
	for i := 0; i < len(r.activeThreads); i++ {
		currentThread := r.activeThreads[i]
		// This call may append to the active and blocked threads lists.
		result := r.runThreadUntilBlockedOrCompleted(currentThread, event)
		if result.Accepted {
			r.reset()
			result.Reset = true
			return result
		}
	}

	// If there are no blocked threads, restart from the beginning of the program.
	if len(r.blockedThreads) == 0 {
		r.reset()
		return Result{Reset: true}
	}

	// No threads accepted for this event.
	return Result{Accepted: false}
}

func (r *Runtime) reset() {
	r.activeThreads = r.activeThreads[:0]
	r.blockedThreads = append(r.blockedThreads[:0], threadState{})
	r.eventCount = 0
}

func (r *Runtime) runThreadUntilBlockedOrCompleted(thread threadState, event Event) Result {
	for {
		bytecode := r.program[thread.programCounter]
		switch bytecode.op {
		case opRead:
			if thread.numConsumed == r.eventCount {
				// Already consumed the event, so block waiting for the next event.
				r.blockedThreads = append(r.blockedThreads, thread)
				return Result{Accepted: false}
			} else if event >= Event(bytecode.arg1) && event <= Event(bytecode.arg2) {
				// Event within the read range, so consume it and continue.
				thread.numConsumed++
				thread.programCounter++
			} else {
				// Event does not match the read range, so this thread dies.
				return Result{Accepted: false}
			}

		case opJump:
			thread.programCounter = int(bytecode.arg1)

		case opFork:
			thread.programCounter = int(bytecode.arg1)
			forkedThread := thread.fork(int(bytecode.arg2))
			r.activeThreads = append(r.activeThreads, forkedThread)

		case opStartCapture:
			thread.startCapture(CaptureId(bytecode.arg1))
			thread.programCounter++

		case opEndCapture:
			thread.endCapture(CaptureId(bytecode.arg1))
			thread.programCounter++

		case opAccept:
			r.activeThreads = r.activeThreads[:0]
			r.blockedThreads = r.blockedThreads[:0]
			return Result{
				Accepted: true,
				Captures: thread.captures,
			}

		default:
			panic("Unrecognized bytecode op")
		}
	}
}

// threadState represents a thread of execution in the virtual machine.
type threadState struct {
	programCounter int
	numConsumed    int
	captures       []Capture
}

// fork spawns a new thread with the same captures as the current thread,
// but at a different instruction in the program.
func (t threadState) fork(programCounter int) threadState {
	var captures []Capture
	if t.captures != nil {
		captures = make([]Capture, len(t.captures))
		copy(captures, t.captures)
	}

	return threadState{
		programCounter: programCounter,
		numConsumed:    t.numConsumed,
		captures:       captures,
	}
}

func (t *threadState) startCapture(captureId CaptureId) {
	// If the program is constructed correctly, there should be at most one start capture
	// per thread, so we don't need to check that a capture with the same ID already exists.
	t.captures = append(t.captures, Capture{
		Id:       captureId,
		StartIdx: t.numConsumed,
	})
}

func (t *threadState) endCapture(captureId CaptureId) {
	// We expect at most a few captures per thread, so it's okay to use a linear search.
	for i := 0; i < len(t.captures); i++ {
		if t.captures[i].Id == captureId {
			t.captures[i].Length = t.numConsumed - t.captures[i].StartIdx
			return
		}
	}
	// This should never happen because a correctly constructed program should
	// always have matching start/end captures.
	panic("Could not find capture")
}
