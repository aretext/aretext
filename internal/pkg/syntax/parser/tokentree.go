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

// Token represents a distinct element in a larger text.
type Token struct {
	StartPos uint64
	EndPos   uint64
	Role     TokenRole
}

// TokenTree represents a collection of tokens.
// It supports efficient lookups by position and "shifting" token positions to account for insertions/deletions.
type TokenTree struct {
	// nodes represents a full binary search tree.
	// For non-full trees, some of the nodes in the slice won't be part of the tree; these nodes have initializedFlag set to zero.
	nodes []treeNode
}

// NewTokenTree constructs a token tree from a set of tokens.
// The tokens must be sorted ascending by start position and non-overlapping.
func NewTokenTree(tokens []Token) *TokenTree {
	type stackItem struct {
		nodeIdx int
		tokens  []Token
	}

	if len(tokens) == 0 {
		return &TokenTree{}
	}

	// Allocate more space than we need to avoid reallocations when new tokens are inserted.
	nodes := make([]treeNode, 2*len(tokens))

	// Construct a balanced binary search tree by recursively building subtrees.
	item := stackItem{0, tokens[:]}
	stack := make([]stackItem, 0, len(tokens))
	stack = append(stack, item)
	for len(stack) > 0 {
		item, stack = stack[len(stack)-1], stack[:len(stack)-1]
		nodeIdx, tokens := item.nodeIdx, item.tokens
		tokenIdx := len(tokens) / 2
		nodes[nodeIdx] = newTreeNodeForToken(tokens[tokenIdx])
		if tokenIdx > 0 {
			stack = append(stack, stackItem{
				nodeIdx: leftChildIdx(nodeIdx),
				tokens:  tokens[0:tokenIdx],
			})
		}
		if tokenIdx+1 < len(item.tokens) {
			stack = append(stack, stackItem{
				nodeIdx: rightChildIdx(nodeIdx),
				tokens:  tokens[tokenIdx+1 : len(tokens)],
			})
		}
	}
	return &TokenTree{nodes}
}

// IterFromPosition returns a token iterator from the specified position.
// If a token overlaps the position, the iterator starts at that token.
// Otherwise, the iterator starts from the next token after the position.
func (t *TokenTree) IterFromPosition(pos uint64) *TokenIter {
	return &TokenIter{
		tree:    t,
		nodeIdx: t.nodeIdxForPos(pos),
	}
}

func (t *TokenTree) nodeIdxForPos(pos uint64) int {
	idx, closestIdxAfter := 0, -1
	for t.isValidNode(idx) {
		root := t.nodes[idx]
		if root.intersects(pos) {
			return idx
		} else if pos < root.startPos() {
			closestIdxAfter = idx
			idx = leftChildIdx(idx)
		} else {
			idx = rightChildIdx(idx)
		}
	}
	return closestIdxAfter
}

func (t *TokenTree) nextNodeIdx(idx int) int {
	// If the current node has a right subtree, the next node is the leftmost node in that subtree.
	if right := rightChildIdx(idx); t.isValidNode(right) {
		return t.leftmostChildInSubtree(right)
	}

	// Traverse up the tree until we find a left child.
	// (If we're in a right child, then we know we've already output all the
	// nodes in the right subtree, which means we've also output the parent,
	// so we need to continue traversing up.)
	for isRightChildIdx(idx) {
		idx = parentIdx(idx)
	}

	// If we're at the root index, we can't go up any further.
	if isRootIdx(idx) {
		return -1
	}

	// We've iterated through all the nodes in the left subtree of the parent,
	// so the parent node comes next.
	return parentIdx(idx)
}

func (t *TokenTree) leftmostChildInSubtree(idx int) int {
	for {
		if left := leftChildIdx(idx); t.isValidNode(left) {
			// Current node has a valid left child, so repeat the search from the left child.
			idx = left
		} else {
			// Current node does NOT have a left child, so it must be the leftmost child.
			return idx
		}
	}
}

func (t *TokenTree) isValidNode(idx int) bool {
	return idx >= 0 && idx < len(t.nodes) && t.nodes[idx].initialized()
}

// TokenIter iterates over tokens.
type TokenIter struct {
	tree    *TokenTree
	nodeIdx int
}

// Next retrieves the next token, if it exists.
func (iter *TokenIter) Next(tok *Token) bool {
	if !iter.tree.isValidNode(iter.nodeIdx) {
		// There are no more tokens to return.
		return false
	}

	*tok = iter.tree.nodes[iter.nodeIdx].token
	iter.nodeIdx = iter.tree.nextNodeIdx(iter.nodeIdx)
	return true
}

// Collect retrieves all tokens from the iterator and returns them as a slice.
func (iter *TokenIter) Collect() []Token {
	result := make([]Token, 0)
	var tok Token
	for iter.Next(&tok) {
		result = append(result, tok)
	}
	return result
}

// Helper methods to calculate indexes in the array representation of the token tree.
// Nodes are stored in pre-order traversal order.
// If indices were 1-indexed, the left child would be idx*2 and the right child would be idx*2+1.
// Since indices are 0-indexed, we need to increment/decrement by one before/after each calculation.
func parentIdx(idx int) int        { return (idx+1)/2 - 1 }
func leftChildIdx(idx int) int     { return (idx+1)*2 - 1 }
func rightChildIdx(idx int) int    { return (idx + 1) * 2 }
func isRootIdx(idx int) bool       { return idx == 0 }
func isRightChildIdx(idx int) bool { return (idx+1)%2 > 0 }

const (
	initializedFlag = 1 << iota
)

type treeNode struct {
	flags int
	token Token
}

func newTreeNodeForToken(token Token) treeNode {
	return treeNode{
		flags: initializedFlag,
		token: token,
	}
}

func (tn *treeNode) initialized() bool {
	return (tn.flags & initializedFlag) > 0
}

func (tn *treeNode) startPos() uint64 {
	return tn.token.StartPos
}

func (tn *treeNode) intersects(pos uint64) bool {
	return pos >= tn.token.StartPos && pos < tn.token.EndPos
}
