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
	treeHeight     uint64
	tokens         []ComputedToken // Only in leaves.
	leftChild      *Computation
	rightChild     *Computation
}

// TODO
func NewComputation(readLength uint64, consumedLength uint64, tokens []ComputedToken) *Computation {
	if consumedLength > readLength {
		panic("Consumed length must be less than or equal to read length")
	}

	for _, tok := range tokens {
		if tok.Offset+tok.Length > consumedLength {
			panic("Token length must be less tha consumed length")
		}
	}

	return &Computation{
		readLength:     readLength,
		consumedLength: consumedLength,
		treeHeight:     1,
		tokens:         tokens,
	}
}

// TODO
func (c *Computation) Append(other *Computation) *Computation {
	if c.treeHeight < other.treeHeight {
		return other.prependSubtree(c)
	} else {
		return c.appendSubtree(other)
	}
}

func (c *Computation) prependSubtree(other *Computation) *Computation {
	if c.leftChild == nil || c.treeHeight <= other.treeHeight {
		return computationFromChildren(other.Append(c.leftChild), c)
	}
	return computationFromChildren(c.leftChild.prependSubtree(other), c.rightChild)
}

func (c *Computation) appendSubtree(other *Computation) *Computation {
	if c.rightChild == nil || c.treeHeight <= other.treeHeight {
		return computationFromChildren(c, other.Append(c.rightChild))
	}
	return computationFromChildren(c.leftChild, c.rightChild.appendSubtree(other))
}

func computationFromChildren(leftChild, rightChild *Computation) *Computation {
	return &Computation{
		readLength:     leftChild.readLength + rightChild.readLength,
		consumedLength: leftChild.consumedLength + rightChild.consumedLength,
		treeHeight:     maxTreeHeight(leftChild, rightChild) + 1,
		leftChild:      leftChild,
		rightChild:     rightChild,
	}
}

func maxTreeHeight(c1, c2 *Computation) uint64 {
	if c1.treeHeight > c2.treeHeight {
		return c1.treeHeight
	} else {
		return c2.treeHeight
	}
}

// TODO
func (c *Computation) LargestSubComputationInRange(readStartPosEquals, readEndPosLessThan uint64) *Computation {
	// TODO: binary search for start pos
	// TODO: search down left spine until computation with readEndPos < readEndPosLessThan
	return nil
}

// TODO
func (c *Computation) TokenIterFromPos(pos uint64) *TokenIter {
	// TODO
	return nil
}
