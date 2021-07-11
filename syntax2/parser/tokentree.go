package parser

// TokenTree represents a set of non-overlapping tokens ordered by start position.
// The tree is immutable; all modifications are made by copying/adding new nodes
// rather than mutating existing nodes.
type TokenTree struct {
	token       Token
	shift       Shift
	minStartPos uint64
	maxEndPos   uint64
	leftChild   *TokenTree
	rightChild  *TokenTree
}

// Insert inserts a new token into the tree.
// The token must have non-zero length and must not overlap existing tokens.
func (t *TokenTree) Insert(token Token) *TokenTree {
	if token.StartPos >= token.EndPos {
		panic("Token length must be positive")
	}
	return t.insertWithShift(token, Shift{})
}

func (t *TokenTree) insertWithShift(token Token, ancestorShift Shift) *TokenTree {
	if t == nil {
		return &TokenTree{
			token:       token,
			shift:       ancestorShift.Negate(), // TODO: explain
			minStartPos: token.StartPos,
			maxEndPos:   token.EndPos,
		}
	}

	shift := ancestorShift.Add(t.shift)
	if token.EndPos <= shift.Resolve(t.token.StartPos) {
		child := t.leftChild.insertWithShift(token, shift)
		return t.withLeftChild(child)
	} else if token.StartPos >= shift.Resolve(t.token.EndPos) {
		child := t.rightChild.insertWithShift(token, shift)
		return t.withRightChild(child)
	} else {
		panic("Token overlaps existing token")
	}
}

// Join combines two trees into a single tree.
// The spans (start of first token to end of last token) of the two trees must not overlap.
func (t *TokenTree) Join(other *TokenTree) *TokenTree {
	if t == nil && other == nil {
		return nil
	} else if other == nil {
		return t
	} else if t == nil {
		return other
	} else if other.shift.Resolve(other.maxEndPos) <= t.shift.Resolve(t.minStartPos) {
		return t.joinBefore(other, Shift{})
	} else if other.shift.Resolve(other.minStartPos) >= t.shift.Resolve(t.maxEndPos) {
		return t.joinAfter(other, Shift{})
	} else {
		panic("Span of other tree overlaps span of this tree")
	}
}

func (t *TokenTree) joinBefore(other *TokenTree, ancestorShift Shift) *TokenTree {
	if t == nil {
		return other.ShiftPositions(ancestorShift.Negate())
	}
	shift := ancestorShift.Add(t.shift)
	child := t.leftChild.joinBefore(other, shift)
	return t.withLeftChild(child)
}

func (t *TokenTree) joinAfter(other *TokenTree, ancestorShift Shift) *TokenTree {
	if t == nil {
		return other.ShiftPositions(ancestorShift.Negate())
	}
	shift := ancestorShift.Add(t.shift)
	child := t.rightChild.joinAfter(other, shift)
	return t.withRightChild(child)
}

// ShiftPositions moves all the tokens in the tree forward or backward in O(1) time.
func (t *TokenTree) ShiftPositions(shift Shift) *TokenTree {
	newTree := *t
	newTree.shift = t.shift.Add(shift)
	return &newTree
}

// IterFromPosition returns a token iterator from the first token ending after the given position.
func (t *TokenTree) IterFromPosition(pos uint64) *TokenIter {
	var shift Shift
	stack := make([]iterStackItem, 0)
	for t != nil {
		shift = shift.Add(t.shift)
		if pos < t.token.StartPos {
			// Position is before this token, so it must be in the left subtree.
			// This node is *after* the left subtree in the in-order traversal,
			// so append it to the stack.
			stack = append(stack, iterStackItem{tree: t, shift: shift})
			t = t.leftChild
		} else if pos >= t.token.EndPos {
			// Position is after this token, so it must be in the right subtree.
			// This node is *before* the right subtree in the in-order traversal,
			// so do NOT append it to the stack.
			t = t.rightChild
		} else {
			// Position intersects this token.  This node is the first one
			// to visit in the in-order traversal, so place it on the top
			// of the stack.
			stack = append(stack, iterStackItem{tree: t, shift: shift})
			break
		}
	}
	return &TokenIter{stack}
}

func (t *TokenTree) withLeftChild(child *TokenTree) *TokenTree {
	newTree := *t
	newTree.leftChild = child
	if child.minStartPos < newTree.minStartPos {
		newTree.minStartPos = child.minStartPos
	}
	return &newTree
}

func (t *TokenTree) withRightChild(child *TokenTree) *TokenTree {
	newTree := *t
	newTree.rightChild = child
	if child.maxEndPos > newTree.maxEndPos {
		newTree.maxEndPos = child.maxEndPos
	}
	return &newTree
}

// TODO
type iterStackItem struct {
	tree  *TokenTree
	shift Shift
}

// TokenIter iterates over tokens.
type TokenIter struct {
	// Stack of tree nodes to visit next.
	// The last element (top of the stack) is the current node.
	stack []iterStackItem
}

// Get retrieves the current token, if it exists.
func (iter *TokenIter) Get(tok *Token) bool {
	if len(iter.stack) == 0 {
		return false
	}

	item := iter.stack[len(iter.stack)-1]
	shift := item.shift
	*tok = item.tree.token
	tok.StartPos = shift.Resolve(tok.StartPos)
	tok.EndPos = shift.Resolve(tok.EndPos)
	return true
}

// Advance moves the iterator to the next token.
// If there are no more tokens, this is a no-op.
func (iter *TokenIter) Advance() {
	if len(iter.stack) == 0 {
		return
	}

	// Pop the current node from the stack,
	// and push all the left children of the current node's right subtree.
	item := iter.stack[len(iter.stack)-1]
	iter.stack = iter.stack[0 : len(iter.stack)-1]
	shift := item.shift
	t := item.tree.rightChild
	for t != nil {
		shift = shift.Add(t.shift)
		iter.stack = append(iter.stack, iterStackItem{tree: t, shift: shift})
		t = t.leftChild
	}
}

// Collect retrieves all tokens from the iterator and returns them as a slice.
func (iter *TokenIter) Collect() []Token {
	result := make([]Token, 0)
	var tok Token
	for iter.Get(&tok) {
		result = append(result, tok)
		iter.Advance()
	}
	return result
}
