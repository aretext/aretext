package parser

// TokenTree represents a set of non-overlapping tokens ordered by start position.
// The tree is immutable; all modifications are made by copying/adding new nodes
// rather than mutating existing nodes.
type TokenTree struct {
	token       Token
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
	return t.insert(token)
}

func (t *TokenTree) insert(token Token) *TokenTree {
	if t == nil {
		return &TokenTree{
			token:       token,
			minStartPos: token.StartPos,
			maxEndPos:   token.EndPos,
		}
	} else if token.EndPos <= t.token.StartPos {
		return t.withLeftChild(t.leftChild.Insert(token))
	} else if token.StartPos >= t.token.EndPos {
		return t.withRightChild(t.rightChild.Insert(token))
	} else {
		panic("Token overlaps existing token")
	}
}

// Join combines two trees into a single tree.
// The spans (start of first token to end of last token) of the two trees must not overlap.
func (t *TokenTree) Join(other *TokenTree) *TokenTree {
	if t == nil && other == nil {
		return nil
	} else if t == nil {
		return other
	} else if other == nil {
		return t
	} else if other.maxEndPos <= t.minStartPos {
		return t.joinBefore(other)
	} else if other.minStartPos >= t.maxEndPos {
		return t.joinAfter(other)
	} else {
		panic("Span of other tree overlaps span of this tree")
	}
}

func (t *TokenTree) joinBefore(other *TokenTree) *TokenTree {
	if t == nil {
		return other
	}
	child := t.leftChild.joinBefore(other)
	return t.withLeftChild(child)
}

func (t *TokenTree) joinAfter(other *TokenTree) *TokenTree {
	if t == nil {
		return other
	}
	child := t.rightChild.joinAfter(other)
	return t.withRightChild(child)
}

// IterFromPosition returns a token iterator from the first token ending after the given position.
func (t *TokenTree) IterFromPosition(pos uint64) *TokenIter {
	stack := make([]*TokenTree, 0)
	for t != nil {
		if pos < t.token.StartPos {
			// Position is before this token, so it must be in the left subtree.
			// This node is *after* the left subtree in the in-order traversal,
			// so append it to the stack.
			stack = append(stack, t)
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
			stack = append(stack, t)
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

// TokenIter iterates over tokens.
type TokenIter struct {
	// Stack of nodes to visit next.
	// The last element (top of the stack) is the current node.
	stack []*TokenTree
}

// Get retrieves the current token, if it exists.
func (iter *TokenIter) Get(tok *Token) bool {
	if len(iter.stack) == 0 {
		return false
	}

	t := iter.stack[len(iter.stack)-1]
	*tok = t.token
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
	t := iter.stack[len(iter.stack)-1].rightChild
	iter.stack = iter.stack[0 : len(iter.stack)-1]
	for t != nil {
		iter.stack = append(iter.stack, t)
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
