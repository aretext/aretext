package parser

type tokenIterStackItem struct {
	startPos uint64
	c        *Computation
}

// TODO
type TokenIter struct {
	stack    []tokenIterStackItem
	tokenIdx int
}

func NewTokenIter(c *Computation, pos uint64) *TokenIter {
	stack := stackForSmallestSubComputationContainingPos(c, pos)
	iter := TokenIter{stack}
	iter.advanceToNextNonEmptyLeaf()
	return iter
}

func stackForSmallestSubComputationContainingPos(c *Computation, pos uint64) []tokenIterStackItem {
	if c == nil || pos >= c.consumedLength {
		return nil
	}

	stack := make([]tokenIterStackItem, 0)
	startPos, endPos := 0, c.consumedLength
	for c != nil {
		if c.leftChild == nil && c.rightChild == nil {
			stack = append(stack, tokenIterStackItem{startPos, computation})
			break
		} else if c.leftChild == nil || pos >= startPos+c.leftChild.consumedLength {
			if c.leftChild != nil {
				startPos += c.leftChild.consumedLength
			}
			c = c.rightChild
		} else {
			if c.rightChild != nil {
				endPos -= c.rightChild.consumedLength
			}
			stack = append(stack, tokenIterStackItem{startPos, computation})
			c = c.leftChild
		}
	}

	return stack
}

func (iter *TokenIter) advanceToNextNonEmptyLeaf() {
	// TODO
}

func (iter *TokenIter) Get(tok *Token) bool {
	return false
}

func (iter *TokenIter) Advance() {
}

func (iter *TokenIter) Collect() []Token {
	// TODO
	return nil
}
