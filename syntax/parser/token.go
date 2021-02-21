package parser

// TokenRole represents the role a token plays.
// This is mainly used for syntax highlighting.
type TokenRole int

const (
	TokenRoleNone = TokenRole(iota)
	TokenRoleOperator
	TokenRoleKeyword
	TokenRoleIdentifier
	TokenRoleNumber
	TokenRoleString
	TokenRoleKey
	TokenRoleComment
	TokenRolePunctuation
)

// Token represents a distinct element in a document.
type Token struct {
	Role     TokenRole
	StartPos uint64
	EndPos   uint64

	// Last position the tokenizer read while constructing the token.
	// This will always be greater than or equal to EndPos.
	LookaheadPos uint64
}

func (t *Token) length() uint64 {
	if t.StartPos > t.EndPos {
		panic("Token has negative length")
	}
	return t.EndPos - t.StartPos
}

func (t *Token) lookaheadLength() uint64 {
	if t.StartPos > t.LookaheadPos || t.EndPos > t.LookaheadPos {
		panic("Token lookahead less than token length")
	}
	return t.LookaheadPos - t.StartPos
}
