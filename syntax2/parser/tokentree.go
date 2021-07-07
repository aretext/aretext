package parser

// TokenTree represents a collection of tokens.
// The zero value represents an empty tree.
// It supports efficient lookups by position and "shifting" token positions to account for insertions/deletions.
type TokenTree struct {
	StartPos     uint64
	EndPos       uint64
	LookaheadPos uint64

	LeftChild  *TokenTree // TODO: only inner
	RightChild *TokenTree // TODO: only inner
	TokenRole  TokenRole  // TODO: only leaf
}

// TODO
func (t *TokenTree) Insert(token Token) *TokenTree {
	return nil
}

// TODO
func (t *TokenTree) Join(other *TokenTree) *TokenTree {
	return nil
}

// TODO
func (t *TokenTree) ShiftPositionsForward(offset uint64) *TokenTree {
	return nil
}

// TODO
func (t *TokenTree) ShiftPositionsBackward(offset uint64) *TokenTree {
	return nil
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
