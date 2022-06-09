package parser

// TokenRole represents the role a token plays in a document,
// as interpreted in a particular syntax language.
type TokenRole int

const (
	TokenRoleNone = TokenRole(iota)
	TokenRoleOperator
	TokenRoleKeyword
	TokenRoleNumber
	TokenRoleString
	TokenRoleComment
)

const (
	TokenRoleCustom1 = TokenRole((1 << 16) + iota)
	TokenRoleCustom2
	TokenRoleCustom3
	TokenRoleCustom4
	TokenRoleCustom5
	TokenRoleCustom6
	TokenRoleCustom7
	TokenRoleCustom8
)

// Token represents a distinct element in a document.
type Token struct {
	Role     TokenRole
	StartPos uint64
	EndPos   uint64
}
