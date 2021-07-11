package parser

// TODO
type ComputedToken struct {
	Offset uint64
	Length uint64
	Role   TokenRole
}

// TODO
type Computation struct {
	ReadLength     uint64
	ConsumedLength uint64
	LeftChild      *Computation
	RightChild     *Computation
	Tokens         []ComputedTokens // Only in leaves.
}
