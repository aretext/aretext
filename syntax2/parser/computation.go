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
	maxChildTreeHeight := leftChild.treeHeight
	if rightChild.treeHeight > maxChildTreeHeight {
		maxChildTreeHeight = rightChild.treeHeight
	}

	return &Computation{
		readLength:     leftChild.readLength + rightChild.readLength,
		consumedLength: leftChild.consumedLength + rightChild.consumedLength,
		treeHeight:     maxChildTreeHeight + 1,
		leftChild:      leftChild,
		rightChild:     rightChild,
	}
}

// TODO
func (c *Computation) LargestSubComputationInRange(readStartPosEquals, readEndPosLessThan uint64) *Computation {
	return c.largestSubComputationInRange(0, c.readLength, readStartPosEquals, readEndPosLessThan)
}

func (c *Computation) largestSubComputationInRange(readStartPos, readEndPos, readStartPosEquals, readEndPosLessThan uint64) *Computation {
	if c.leftChild == nil && c.rightChild == nil {
		return nil
	}

	if c.leftChild == nil {
		return c.rightChild.largestSubComputationInRange(
			readStartPos,
			readEndPos,
			readStartPosEquals,
			readEndPosLessThan,
		)
	}

	if c.rightChild == nil {
		return c.leftChild.largestSubComputationInRange(
			readStartPos,
			readEndPos,
			readStartPosEquals,
			readEndPosLessThan,
		)
	}

	if readStartPos != readStartPosEquals {
		if readStartPosEquals < readStartPos+c.leftChild.readLength {
			return c.leftChild.largestSubComputationInRange(
				readStartPos,
				readEndPos-c.rightChild.readLength,
				readStartPosEquals,
				readEndPosLessThan,
			)
		} else {
			return c.rightChild.largestSubComputationInRange(
				readStartPos+c.leftChild.readLength,
				readEndPos,
				readStartPosEquals,
				readEndPosLessThan,
			)
		}
	}

	if readEndPos >= readEndPosLessThan {
		return c.leftChild.largestSubComputationInRange(
			readStartPos,
			readEndPos-c.rightChild.readLength,
			readStartPosEquals,
			readEndPosLessThan,
		)
	}

	return c
}

// TODO
func (c *Computation) TokenIterFromPos(pos uint64) *TokenIter {
	// TODO
	return nil
}
