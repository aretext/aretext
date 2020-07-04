package breaks

// circularRuneBuffer is a double-ended FIFO queue for runes.
type circularRuneBuffer struct {
	slots    []rune
	capacity int
	headIdx  int
	length   int
}

// newCircularRuneBuffer constructs an empty buffer.
func newCircularRuneBuffer(capacity int) *circularRuneBuffer {
	return &circularRuneBuffer{
		slots:    make([]rune, capacity),
		capacity: capacity,
		headIdx:  0,
		length:   0,
	}
}

// canWrite checks if there is space in the buffer for at least one additional rune.
func (b *circularRuneBuffer) canWrite() bool {
	return b.length < b.capacity
}

// write appends a rune to the queue.
// It panics if the buffer is full.
func (b *circularRuneBuffer) write(r rune) {
	if b.length+1 > b.capacity {
		panic("Cannot write to a full buffer")
	}

	idx := (b.headIdx + b.length) % b.capacity
	b.slots[idx] = r
	b.length++
}

// canConsume checks if there is a rune in the buffer to consume.
func (b *circularRuneBuffer) canConsume() bool {
	return b.length > 0
}

// consume removes and returns the oldest rune in the buffer.
// It panics if the buffer is empty.
func (b *circularRuneBuffer) consume() rune {
	if b.length < 1 {
		panic("Cannot consume from an empty buffer")
	}

	r := b.slots[b.headIdx]
	b.headIdx = (b.headIdx + 1) % b.capacity
	b.length--
	return r
}

// peekAtIdx returns a copy of the rune at an index in the buffer.
// The first index 0 is the oldest rune in the buffer.
// If no rune exists at the index, it returns false.
func (b *circularRuneBuffer) peekAtIdx(i int) (rune, bool) {
	if i < 0 || i >= b.length {
		return '\x00', false
	}

	idx := (b.headIdx + i) % b.capacity
	return b.slots[idx], true
}
