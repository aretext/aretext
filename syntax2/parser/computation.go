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

func NewComputation(readLen uint64, consumedLen uint64, tokens []ComputedToken) *Computation {
	// TODO
	return nil

}

// TODO
func (c *Computation) Append(other *Computation) *Computation {
	// TODO
	return nil
}

// TODO
func (c *Computation) Search(readStartPosEquals, readEndPosLessThan uint64) *Computation {
	// TODO
	return nil
}

// TODO
func (c *Computation) TokenIterFromPos(pos uint64) *TokenIter {
	// TODO
	return nil
}
