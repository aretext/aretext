package parser

// TODO
type ComputedToken struct {
	Offset uint64
	Length uint64
	Role   TokenRole
}

// TODO
type Computation struct {
	ReadLength     uint64
	ConsumedLength uint64
	Tokens         []ComputedTokens // Only in leaves.
	leftChild      *Computation
	rightChild     *Computation
}

// TODO
func (c *Computation) Append(other *Computation) *Computation {
	// TODO
	return nil
}

// TODO
func (c *Computation) Lookup( /* criteria? */ ) *Computation {
	// TODO
	return nil
}

// TODO
func (c *Computation) TokenIterFromPos(pos uint64) *TokenIter {
	// TODO
	return nil
}
