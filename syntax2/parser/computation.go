package parser

// ComputedToken is a token recognized by a computation.
type ComputedToken struct {
	// Offset is the token's start position,
	// defined relative to the computation's start position.
	Offset uint64
	Length uint64
	Role   TokenRole
}

// Computation is a result produced by a tokenizer.
// Computations are composable, so part of one computation
// can be re-used when re-parsing an edited text.
type Computation struct {
	readLength     uint64
	consumedLength uint64
	treeHeight     uint64
	tokens         []ComputedToken // Only in leaves.
	leftChild      *Computation
	rightChild     *Computation
}

// NewComputation constructs a computation.
// readLength is the number of runes read by the tokenizer,
// and consumedLength is the number of runes consumed by the tokenizer.
// The tokens slice contains any tokens recognized by the tokenizer;
// these must have non-zero length, be ordered sequentially by start position,
// and be non-overlapping.
func NewComputation(readLength uint64, consumedLength uint64, tokens []ComputedToken) *Computation {
	if consumedLength > readLength {
		panic("Consumed length must be less than or equal to read length")
	}

	var lastEndPos uint64
	for _, tok := range tokens {
		if tok.Length == 0 {
			panic("Token must have non-zero length")
		}

		if tok.Offset < lastEndPos {
			panic("Token must be sequential and non-overlapping")
		}

		tokEndPos := tok.Offset + tok.Length
		if tokEndPos > consumedLength {
			panic("Token length must be less than consumed length")
		}

		lastEndPos = tokEndPos
	}

	return &Computation{
		readLength:     readLength,
		consumedLength: consumedLength,
		treeHeight:     1,
		tokens:         tokens,
	}
}

// ReadLength returns the number of runes read to produce this computation.
func (c *Computation) ReadLength() uint64 {
	if c == nil {
		return 0
	} else {
		return c.readLength
	}
}

// ConsumedLength returns the number of runes consumed to produce this computation.
func (c *Computation) ConsumedLength() uint64 {
	if c == nil {
		return 0
	} else {
		return c.consumedLength
	}
}

// TreeHeight returns the height of the computation tree.
func (c *Computation) TreeHeight() uint64 {
	if c == nil {
		return 0
	} else {
		return c.treeHeight
	}
}

// Append appends one computation after another computation.
// The positions of the computations and tokens in the second computation
// are "shifted" to start immediately after the end (consumed length) of
// the first computation.
func (c *Computation) Append(other *Computation) *Computation {
	if c == nil {
		return other
	} else if other == nil {
		return c
	}

	// This is the AVL join algorithm from
	// Blelloch, G. E., Ferizovic, D., & Sun, Y. (2016). Just join for parallel ordered sets.
	// In Proceedings of the 28th ACM Symposium on Parallelism in Algorithms and Architectures.
	h1, h2 := c.TreeHeight(), other.TreeHeight()
	if h1 == h2 {
		return computationFromChildren(c, other)
	} else if h1 < h2 {
		return other.prependSubtree(c)
	} else {
		return c.appendSubtree(other)
	}
}

// prependSubtree inserts a computation *before* a given computation,
// rebalancing the tree if necessary (AVL balance invariant).
// This assumes that both computations are non-nil.
func (c *Computation) prependSubtree(other *Computation) *Computation {
	if c.leftChild.TreeHeight() <= other.TreeHeight()+1 {
		// Insert the new tree as a sibling of a left child with approximately the same height.
		newLeft := computationFromChildren(other, c.leftChild)
		if newLeft.TreeHeight() <= c.rightChild.TreeHeight()+1 {
			// The new tree already satisfies the AVL balance invariant.
			return computationFromChildren(newLeft, c.rightChild)
		} else {
			// The new tree violates the AVL balance invariant.
			// Double-rotate to restore balance.
			return computationFromChildren(newLeft.rotateLeft(), c.rightChild).rotateRight()
		}
	}

	// Recursively search for a sibling with approximately the same height as the inserted subtree.
	newLeft := c.leftChild.prependSubtree(other)
	newRoot := computationFromChildren(newLeft, c.rightChild)
	if newLeft.TreeHeight() <= c.rightChild.TreeHeight()+1 {
		// The new tree already satisfies the AVL balance invariant.
		return newRoot
	} else {
		// The new tree violates the AVL balance invariant.
		// Rotate to restore balance.
		return newRoot.rotateRight()
	}
}

// appendSubtree inserts a computation *after* a given computation,
// rebalancing the tree if necessary (AVL balance invariant).
// This assumes that both computations are non-nil.
func (c *Computation) appendSubtree(other *Computation) *Computation {
	if c.rightChild.TreeHeight() <= other.TreeHeight()+1 {
		// Insert the new tree as a sibling of a right child with approximately the same height.
		newRight := computationFromChildren(c.rightChild, other)
		if newRight.TreeHeight() <= c.leftChild.TreeHeight()+1 {
			// The new tree already satisfies the AVL balance invariant.
			return computationFromChildren(c.leftChild, newRight)
		} else {
			// The new tree violates the AVL balance invariant.
			// Double-rotate to restore balance.
			return computationFromChildren(c.leftChild, newRight.rotateRight()).rotateLeft()
		}
	}

	// Recursively search for a sibling with approximately the same height as the inserted subtree.
	newRight := c.rightChild.appendSubtree(other)
	newRoot := computationFromChildren(c.leftChild, newRight)
	if newRight.TreeHeight() <= c.leftChild.TreeHeight()+1 {
		// The new tree already satisfies the AVL balance invariant.
		return newRoot
	} else {
		// The new tree violates the AVL balance invariant.
		// Rotate to restore balance.
		return newRoot.rotateLeft()
	}
}

func (c *Computation) rotateLeft() *Computation {
	if c == nil || c.rightChild == nil {
		// Can't rotate left for an empty tree or tree without a right child.
		return c
	}

	//    [x]                [y']
	//   /   \              /   \
	//  [q]  [y]    ==>   [x']   [s]
	//      /   \        /   \
	//    [r]   [s]     [q]  [r]
	x := c
	y := x.rightChild
	q := x.leftChild
	r := y.leftChild
	s := y.rightChild

	if r == nil && s == nil {
		// If y is a leaf, then we can't rotate it into an inner node
		// without losing information about the original computation,
		// so copy y into the leaf node position.
		// This does not change the height of the resulting tree.
		s = y
	}

	return computationFromChildren(computationFromChildren(q, r), s)
}

func (c *Computation) rotateRight() *Computation {
	if c == nil || c.leftChild == nil {
		// Can't rotate right for an empty tree or tree without a left child.
		return c
	}

	//       [x]                [y']
	//      /   \              /   \
	//     [y]  [s]    ==>   [q]   [x']
	//    /   \                    /   \
	//  [q]   [r]                [r]   [s]
	x := c
	y := x.leftChild
	q := y.leftChild
	r := y.rightChild
	s := x.rightChild

	if q == nil && r == nil {
		// If y is a leaf, then we can't rotate it into an inner node
		// without losing information about the original computation,
		// so copy y into the leaf node position.
		// This does not change the height of the resulting tree.
		q = y
	}

	return computationFromChildren(q, computationFromChildren(r, s))
}

func computationFromChildren(leftChild, rightChild *Computation) *Computation {
	if leftChild == nil && rightChild == nil {
		return nil
	}

	maxChildTreeHeight := leftChild.TreeHeight()
	if rightChild.TreeHeight() > maxChildTreeHeight {
		maxChildTreeHeight = rightChild.TreeHeight()
	}

	// Right child starts reading after last character consumed by left child.
	maxReadLength := leftChild.ConsumedLength() + rightChild.ReadLength()
	if leftChild.ReadLength() > maxReadLength {
		maxReadLength = leftChild.ReadLength()
	}

	return &Computation{
		readLength:     maxReadLength,
		consumedLength: leftChild.ConsumedLength() + rightChild.ConsumedLength(),
		treeHeight:     maxChildTreeHeight + 1,
		leftChild:      leftChild,
		rightChild:     rightChild,
	}
}

// LargestSubComputationInRange returns the largest sub-computation with a read range
// contained within the range.
// This is used to find a re-usable computation that is still valid after an edit.
// A computation is considered *invalid* if it read some text that was edited,
// so if the computation did *not* read any edited text, it's definitely still valid.
func (c *Computation) LargestSubComputationInRange(rangeStartPos, rangeEndPos uint64) *Computation {
	return c.largestSubComputationInRange(0, c.readLength, rangeStartPos, rangeEndPos)
}

func (c *Computation) largestSubComputationInRange(readStartPos, readEndPos, rangeStartPos, rangeEndPos uint64) *Computation {

	// First, search until we find a sub-computation with the requested start position.
	if readStartPos != rangeStartPos {
		if c.leftChild == nil && c.rightChild == nil {
			return nil
		} else if c.leftChild == nil {
			// Right child has no sibling, so there's only one direction to search.
			return c.rightChild.largestSubComputationInRange(
				readStartPos,
				readEndPos,
				rangeStartPos,
				rangeEndPos,
			)
		} else if c.rightChild == nil {
			// Left child has no sibling, so there's only one direction to search.
			return c.leftChild.largestSubComputationInRange(
				readStartPos,
				readEndPos,
				rangeStartPos,
				rangeEndPos,
			)
		} else if rangeStartPos < readStartPos+c.leftChild.consumedLength {
			return c.leftChild.largestSubComputationInRange(
				readStartPos,
				readStartPos+c.leftChild.readLength,
				rangeStartPos,
				rangeEndPos,
			)
		} else {
			// Right child starts reading after last character consumed by left child.
			newReadStartPos := readStartPos + c.leftChild.consumedLength
			newReadEndPos := newReadStartPos + c.rightChild.readLength
			return c.rightChild.largestSubComputationInRange(
				newReadStartPos,
				newReadEndPos,
				rangeStartPos,
				rangeEndPos,
			)
		}
	}

	// Keep searching smaller and smaller sub-computations with the requested start position
	// until we find one that didn't read past the end position.
	if readEndPos > rangeEndPos {
		if c.leftChild == nil && c.rightChild == nil {
			return nil
		} else if c.leftChild == nil {
			// Right child has no sibling, so there's only one direction to search.
			return c.rightChild.largestSubComputationInRange(
				readStartPos,
				readEndPos,
				rangeStartPos,
				rangeEndPos,
			)
		} else if c.rightChild == nil {
			// Left child has no sibling, so there's only one direction to search.
			return c.leftChild.largestSubComputationInRange(
				readStartPos,
				readEndPos,
				rangeStartPos,
				rangeEndPos,
			)
		} else {
			return c.leftChild.largestSubComputationInRange(
				readStartPos,
				readStartPos+c.leftChild.readLength,
				rangeStartPos,
				rangeEndPos,
			)
		}
	}

	return c
}

// TokensIntersectingRange returns tokens that overlap the interval [startPos, endPos)
func (c *Computation) TokensIntersectingRange(startPos, endPos uint64) []Token {
	if c == nil {
		return nil
	}

	var result []Token

	type stackItem struct {
		offset uint64
		c      *Computation
	}
	item := stackItem{offset: 0, c: c}
	stack := []stackItem{item}

	for len(stack) > 0 {
		item, stack = stack[len(stack)-1], stack[0:len(stack)-1]
		offset, c := item.offset, item.c

		if endPos <= offset || offset+c.consumedLength <= startPos {
			// The range doesn't intersect this computation or any of its children.
			continue
		}

		// Find all tokens from this computation that intersect the range
		// (only leaf nodes have tokens).
		for _, computedToken := range c.tokens {
			tok := Token{
				StartPos: offset + computedToken.Offset,
				EndPos:   offset + computedToken.Offset + computedToken.Length,
				Role:     computedToken.Role,
			}
			if !(endPos <= tok.StartPos || startPos >= tok.EndPos) {
				result = append(result, tok)
			}
		}

		// Add tokens from the right child, if it exists.
		// Push this onto the stack first so tokens are added
		// AFTER tokens from the left child.
		if c.rightChild != nil {
			newOffset := offset
			if c.leftChild != nil {
				newOffset += c.leftChild.consumedLength
			}
			stack = append(stack, stackItem{
				offset: newOffset,
				c:      c.rightChild,
			})
		}

		// Add tokens from the left child, if it exists.
		if c.leftChild != nil {
			stack = append(stack, stackItem{
				offset: offset,
				c:      c.leftChild,
			})
		}
	}

	return result
}
