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
	StartPos uint64
	EndPos   uint64
	Role     TokenRole
}

// Edit represents a change to a document.
type Edit struct {
	Pos         uint64 // Position of the first character inserted/deleted.
	NumInserted uint64
	NumDeleted  uint64
}

func (edit Edit) applyToPosition(pos uint64) uint64 {
	if pos >= edit.Pos {
		if updatedPos := pos + edit.NumInserted; updatedPos >= pos {
			pos = updatedPos
		} else {
			pos = uint64(0xFFFFFFFFFFFFFFFF) // overflow
		}

		if updatedPos := pos - edit.NumDeleted; updatedPos <= pos {
			pos = updatedPos
		} else {
			pos = 0 // underflow
		}

		if pos < edit.Pos {
			pos = edit.Pos
		}
	}
	return pos
}

// TokenTree represents a collection of tokens.
// It supports efficient lookups by position and "shifting" token positions to account for insertions/deletions.
type TokenTree struct {
	// nodes represents a full binary search tree.
	// For non-full trees, some of the nodes in the slice won't be part of the tree; these nodes have the initialized flag set to false.
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

// ShiftPositionsAfterEdit adjusts the positions of tokens after an insertion/deletion.
func (t *TokenTree) ShiftPositionsAfterEdit(edit Edit) {
	idx := t.nodeIdxForPos(edit.Pos)
	if t.isValidNode(idx) && t.nodes[idx].intersects(edit.Pos) {
		t.nodes[idx].applyEdit(edit)
		idx = t.nextNodeIdx(idx)
	}

	applyNextEdit := true
	for t.isValidNode(idx) {
		if applyNextEdit {
			t.nodes[idx].applyEdit(edit)
			if right := rightChildIdx(idx); t.isValidNode(right) {
				t.nodes[right].lazyEdits = append(t.nodes[right].lazyEdits, edit)
			}
		}
		applyNextEdit = isLeftChildIdx(idx)
		idx = parentIdx(idx)
	}
}

func (t *TokenTree) nodeIdxForPos(pos uint64) int {
	idx, closestIdxAfter := 0, -1
	for t.isValidNode(idx) {
		t.propagateLazyEdits(idx)
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
		t.propagateLazyEdits(idx)
		if left := leftChildIdx(idx); t.isValidNode(left) {
			// Current node has a valid left child, so repeat the search from the left child.
			idx = left
		} else {
			// Current node does NOT have a left child, so it must be the leftmost child.
			return idx
		}
	}
}

func (t *TokenTree) propagateLazyEdits(idx int) {
	edits := t.nodes[idx].applyLazyEdits()
	if left := leftChildIdx(idx); t.isValidNode(left) {
		t.nodes[left].lazyEdits = append(t.nodes[left].lazyEdits, edits...)
	}
	if right := rightChildIdx(idx); t.isValidNode(right) {
		t.nodes[right].lazyEdits = append(t.nodes[right].lazyEdits, edits...)
	}
}

func (t *TokenTree) isValidNode(idx int) bool {
	return idx >= 0 && idx < len(t.nodes) && t.nodes[idx].initialized
}

// TokenIter iterates over tokens.
// Iterator operations are NOT thread-safe because they can mutate the tree (applying lazy edits).
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
func isLeftChildIdx(idx int) bool  { return idx > 0 && (idx+1)%2 == 0 }
func isRightChildIdx(idx int) bool { return idx > 0 && (idx+1)%2 > 0 }

type treeNode struct {
	initialized bool
	token       Token
	lazyEdits   []Edit
}

func newTreeNodeForToken(token Token) treeNode {
	return treeNode{
		initialized: true,
		token:       token,
	}
}

func (tn *treeNode) startPos() uint64 {
	return tn.token.StartPos
}

func (tn *treeNode) endPos() uint64 {
	return tn.token.EndPos
}

func (tn *treeNode) intersects(pos uint64) bool {
	// For zero-length tokens, intersect for the position at the start of the token.
	// For all other tokens, intersect if the position is in the range [startPos, endPos).
	startPos, endPos := tn.startPos(), tn.endPos()
	return (startPos == endPos && pos == startPos) || (pos >= startPos && pos < endPos)
}

func (tn *treeNode) applyLazyEdits() []Edit {
	edits := tn.lazyEdits
	tn.lazyEdits = nil
	for _, edit := range edits {
		tn.applyEdit(edit)
	}
	return edits
}

func (tn *treeNode) applyEdit(edit Edit) {
	// It is possible for deletions to produce tokens with zero length (start pos == end pos).
	tn.token.StartPos = edit.applyToPosition(tn.token.StartPos)
	tn.token.EndPos = edit.applyToPosition(tn.token.EndPos)
}
