package parser

// TODO
type ComputedToken struct {
	Offset uint64
	Length uint64
	Role   TokenRole
}

// TODO
type Computation struct {
	readLength     uint64
	consumedLength uint64
	tokens         []ComputedToken // Only in leaves.
	leftChild      *Computation
	rightChild     *Computation
}

func NewComputation(readLength uint64, consumedLength uint64, tokens []ComputedToken) *Computation {
	if consumedLength > readLength {
		panic("Consumed length must be less than or equal to read length")
	}

	for _, tok := range tokens {
		if tok.offset+tok.Length > consumedLength {
			panic("Token length must be less tha consumed length")
		}
	}

	return &Computation{
		readLength:     readLength,
		consumedLength: consumedLength,
		tokens:         tokens,
	}
}

// TODO
func (c *Computation) Append(other *Computation) *Computation {
	return &Computation{
		readLength:     c.readLength + other.readLength,
		consumedLength: c.consumedLength + other.consumedLength,
		leftChild:      c,
		rightChild:     other,
	}
}

// TODO
func (c *Computation) LargestSubComputationInRange(readStartPosEquals, readEndPosLessThan uint64) *Computation {
	// TODO
	return nil
}

// TODO
func (c *Computation) TokenIterFromPos(pos uint64) *TokenIter {
	// TODO
	return nil
}
