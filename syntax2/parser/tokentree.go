package parser

// TokenTree represents a set of non-overlapping tokens ordered by start position.
type TokenTree struct {
	StartPos     uint64
	EndPos       uint64
	LookaheadPos uint64
	TokenRole    TokenRole // TODO: only leaf

	MinStartPos uint64
	LeftChild   *TokenTree // TODO: only inner
	RightChild  *TokenTree // TODO: only inner
}

// TODO
func (t *TokenTree) Insert(token Token) *TokenTree {
	if t == nil {
		return &TokenTree{
			MinStartPos:  token.StartPos,
			StartPos:     token.StartPos,
			EndPos:       token.EndPos,
			LookaheadPos: token.LookaheadPos,
			TokenRole:    token.Role,
		}
	}

	if token.StartPos < t.StartPos {
		return &TokenTree{
			MinStartPos:  minUint64(token.StartPos, t.MinStartPos),
			StartPos:     t.StartPos,
			EndPos:       t.EndPos,
			LookaheadPos: t.LookaheadPos,
			TokenRole:    t.TokenRole,
			LeftChild:    t.LeftChild.Insert(token),
		}
	} else if token.StartPos > t.StartPos {
		return &TokenTree{
			MinStartPos:  t.MinStartPos,
			StartPos:     t.StartPos,
			EndPos:       t.EndPos,
			LookaheadPos: t.LookaheadPos,
			TokenRole:    t.TokenRole,
			RightChild:   t.RightChild.Insert(token),
		}
	} else {
		panic("Cannot insert a token with the same start position as an existing token")
	}
}

// IterFromPosition returns a token iterator from the first token ending after the given position.
func (t *TokenTree) IterFromPosition(pos uint64) *TokenIter {
	// TODO
	return nil
}

// TokenIter iterates over tokens.
type TokenIter struct {
	// TODO
}

// Get retrieves the current token, if it exists.
func (iter *TokenIter) Get(tok *Token) bool {
	// TODO
	return false
}

// Advance moves the iterator to the next token.
// If there are no more tokens, this is a no-op.
func (iter *TokenIter) Advance() {
	// TODO
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
