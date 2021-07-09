package parser

// TokenTree represents a set of non-overlapping tokens ordered by start position.
// The tree is immutable; all modifications are made by copying/adding new nodes
// rather than mutating existing nodes.
type TokenTree struct {
	Token       Token // Token.StartPos is the sort key.
	MinStartPos uint64
	LeftChild   *TokenTree
	RightChild  *TokenTree
}

// Insert inserts a new token into the tree.
// The token must have non-zero length and must not overlap existing tokens.
func (t *TokenTree) Insert(token Token) *TokenTree {
	t.validateNewToken(token)
	if t == nil {
		return &TokenTree{Token: token}
	} else if token.StartPos < t.Token.StartPos {
		return t.withLeftChild(t.LeftChild.Insert(token))
	} else {
		return t.withRightChild(t.RightChild.Insert(token))
	}
}

func (t *TokenTree) validateNewToken(token Token) {
	if token.StartPos >= token.EndPos {
		panic("Token length must be positive")
	}
	if token.EndPos > token.LookaheadPos {
		panic("Token lookahead must be greater than or equal to token length")
	}
	if !(t == nil || token.EndPos <= t.Token.StartPos || token.StartPos >= t.Token.EndPos) {
		panic("Token overlaps existing token")
	}
}

func (t *TokenTree) withLeftChild(child *TokenTree) *TokenTree {
	minStartPos := t.MinStartPos
	if child.Token.StartPos < minStartPos {
		minStartPos = child.Token.StartPos
	}
	return &TokenTree{
		MinStartPos: minStartPos,
		Token:       t.Token,
		LeftChild:   child,
		RightChild:  t.RightChild,
	}
}

func (t *TokenTree) withRightChild(child *TokenTree) *TokenTree {
	return &TokenTree{
		MinStartPos: t.MinStartPos,
		Token:       t.Token,
		LeftChild:   t.LeftChild,
		RightChild:  child,
	}
}

// IterFromPosition returns a token iterator from the first token ending after the given position.
func (t *TokenTree) IterFromPosition(pos uint64) *TokenIter {
	stack := make([]*TokenTree, 0)
	for t != nil {
		// Since tokens are non-overlapping and have non-zero length,
		// the sort order by StartPos is the same as the sort order by EndPos.
		// This means we can use EndPos as a sort key here even though
		// we used StartPos as the sort key for insertions.
		if pos < t.Token.EndPos {
			// TODO: explain this
			stack = append(stack, t)
			t = t.LeftChild
		} else if pos >= t.Token.EndPos {
			// TODO: explain this
			if t.RightChild == nil {
				stack = append(stack, t)
				break
			}
			t = t.RightChild
		} else {
			// TODO: explain this
			stack = append(stack, t)
			break
		}
	}
	return &TokenIter{stack}
}

// TokenIter iterates over tokens.
type TokenIter struct {
	// TODO: explain what the stack represents
	stack []*TokenTree
}

// Get retrieves the current token, if it exists.
func (iter *TokenIter) Get(tok *Token) bool {
	if len(iter.stack) == 0 {
		return false
	}

	t := iter.stack[len(iter.stack)-1]
	*tok = t.Token
	return true
}

// Advance moves the iterator to the next token.
// If there are no more tokens, this is a no-op.
func (iter *TokenIter) Advance() {
	if len(iter.stack) == 0 {
		return
	}

	// TODO: explain this
	t := iter.stack[len(iter.stack)-1].RightChild
	iter.stack = iter.stack[0 : len(iter.stack)-1]
	for t != nil {
		iter.stack = append(iter.stack, t)
		t = t.LeftChild
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
