package parser

// TokenTree represents a set of non-overlapping tokens ordered by start position.
// The tree is immutable; all modifications are made by copying/adding new nodes
// rather than mutating existing nodes.
type TokenTree struct {
	StartPos     uint64
	EndPos       uint64
	LookaheadPos uint64
	TokenRole    TokenRole

	MinStartPos uint64
	LeftChild   *TokenTree // TODO: only inner
	RightChild  *TokenTree // TODO: only inner
}

// Insert inserts a new token into the tree.
// The token must have non-zero length and must not overlap existing tokens.
func (t *TokenTree) Insert(token Token) *TokenTree {
	if token.StartPos >= token.EndPos {
		panic("Token length must be positive")
	}
	if token.EndPos >= token.LookaheadPos {
		panic("Token lookahead must be greater than token length")
	}
	if !(token.EndPos <= t.StartPos || token.StartPos >= t.EndPos) {
		panic("Token overlaps existing token")
	}

	if t == nil {
		return treeFromToken(token)
	} else if token.StartPos < t.StartPos {
		return t.withLeftChild(t.LeftChild.Insert(token))
	} else if token.StartPos > t.StartPos {
		return t.withRightChild(t.RightChild.Insert(token))
	} else {
		panic("Cannot insert a token with the same start position as an existing token")
	}
}

func treeFromToken(token Token) *TokenTree {
	return &TokenTree{
		MinStartPos:  token.StartPos,
		StartPos:     token.StartPos,
		EndPos:       token.EndPos,
		LookaheadPos: token.LookaheadPos,
		TokenRole:    token.Role,
	}
}

func (t *TokenTree) withLeftChild(child *TokenTree) *TokenTree {
	minStartPos := t.MinStartPos
	if child.StartPos < minStartPos {
		minStartPos = child.StartPos
	}
	return &TokenTree{
		MinStartPos:  minStartPos,
		StartPos:     t.StartPos,
		EndPos:       t.EndPos,
		LookaheadPos: t.LookaheadPos,
		TokenRole:    t.TokenRole,
		LeftChild:    child,
	}
}

func (t *TokenTree) withRightChild(child *TokenTree) *TokenTree {
	return &TokenTree{
		MinStartPos:  t.MinStartPos,
		StartPos:     t.StartPos,
		EndPos:       t.EndPos,
		LookaheadPos: t.LookaheadPos,
		TokenRole:    t.TokenRole,
		RightChild:   child,
	}
}

// IterFromPosition returns a token iterator from the first token ending after the given position.
func (t *TokenTree) IterFromPosition(pos uint64) *TokenIter {
	iter := t.buildIter(pos)
	advanceIterEndPastPos(iter, pos)
	return iter
}

func (t *TokenTree) buildIter(pos uint64) *TokenIter {
	stack := make([]*TokenTree, 0)
	for t != nil {
		if pos < t.StartPos {
			stack = append(stack, t)
			t = t.LeftChild
		} else if pos > t.StartPos {
			if t.RightChild == nil {
				stack = append(stack, t)
				break
			}
			t = t.RightChild
		} else {
			stack = append(stack, t)
			break
		}
	}
	return &TokenIter{stack}
}

func advanceIterEndPastPos(iter *TokenIter, pos uint64) {
	var tok *Token
	for iter.Get(tok) {
		if tok.EndPos > pos {
			break
		}
		iter.Advance()
	}
}

// TokenIter iterates over tokens.
type TokenIter struct {
	stack []*TokenTree
}

// Get retrieves the current token, if it exists.
func (iter *TokenIter) Get(tok *Token) bool {
	if len(iter.stack) == 0 {
		return false
	}

	t := iter.stack[len(iter.stack)-1]
	*tok = Token{
		StartPos:     t.StartPos,
		EndPos:       t.EndPos,
		LookaheadPos: t.LookaheadPos,
		Role:         t.TokenRole,
	}
	return true
}

// Advance moves the iterator to the next token.
// If there are no more tokens, this is a no-op.
func (iter *TokenIter) Advance() {
	if len(iter.stack) == 0 {
		return
	}

	iter.stack = iter.stack[0 : len(iter.stack)-1]
	if len(iter.stack) == 0 {
		return
	}

	t := iter.stack[len(iter.stack)-1].RightChild
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
