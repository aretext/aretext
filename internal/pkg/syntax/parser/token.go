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
	TokenRoleComment
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
