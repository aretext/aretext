package parser

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
		node := newTreeNodeForToken(tokens[tokenIdx])
		nodes[nodeIdx] = node

		if !isRootIdx(nodeIdx) {
			parent := &nodes[parentIdx(nodeIdx)]
			if parent.maxLookaheadPos < node.maxLookaheadPos {
				parent.maxLookaheadPos = node.maxLookaheadPos
			}
		}

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

// iterFromFirstAffected returns a token iterator from the first token that could be affected by an edit.
// A token could be affected by an edit if the edit occurred in the half-open interval from the token's start position and its lookahead position.
func (t *TokenTree) iterFromFirstAffected(editPos uint64) *TokenIter {
	return &TokenIter{
		tree:    t,
		nodeIdx: t.firstAffectedIdx(editPos),
	}
}

// insertToken adds a new token to the tree.
func (t *TokenTree) insertToken(tok Token) {
	if len(t.nodes) == 0 {
		t.nodes = append(t.nodes, newTreeNodeForToken(tok))
		return
	}

	idx := 0
	for t.isValidNode(idx) {
		t.propagateLazyEdits(idx)
		node := t.nodes[idx]
		if tok.StartPos < node.startPos() {
			idx = leftChildIdx(idx)
		} else {
			idx = rightChildIdx(idx)
		}
	}

	for idx >= len(t.nodes) {
		t.nodes = append(t.nodes, treeNode{})
	}

	t.nodes[idx] = newTreeNodeForToken(tok)
	t.recalculateMaxLookaheadFromIdxToRoot(idx)
}

// shiftPositionsAfterEdit adjusts the positions of tokens after an insertion/deletion.
func (t *TokenTree) shiftPositionsAfterEdit(edit Edit) {
	idx := t.nodeIdxForPos(edit.Pos)
	if t.isValidNode(idx) && t.nodes[idx].intersects(edit.Pos) {
		t.nodes[idx].applyEdit(edit)
		t.recalculateMaxLookaheadFromIdxToRoot(idx)
		idx = t.nextNodeIdx(idx)
	}

	leafIdx := idx
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
	t.recalculateMaxLookaheadFromIdxToRoot(leafIdx)
}

func (t *TokenTree) isValidNode(idx int) bool {
	return idx >= 0 && idx < len(t.nodes) && t.nodes[idx].initialized
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

func (t *TokenTree) firstAffectedIdx(editPos uint64) int {
	firstAffectedStart, firstAffectedIdx := uint64(0xFFFFFFFFFFFFFFFF), -1
	idx := 0
	for t.isValidNode(idx) {
		t.propagateLazyEdits(idx)
		node := t.nodes[idx]
		startPos := node.startPos()
		if startPos < firstAffectedStart && startPos <= editPos && editPos < node.lookaheadPos() {
			firstAffectedStart = startPos
			firstAffectedIdx = idx
		}

		if left := leftChildIdx(idx); t.isValidNode(left) && editPos < t.nodes[left].maxLookaheadPos {
			idx = left
		} else {
			idx = rightChildIdx(idx)
		}
	}

	return firstAffectedIdx
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

func (t *TokenTree) rightmostChildInSubtree(idx int) int {
	for {
		t.propagateLazyEdits(idx)
		if right := rightChildIdx(idx); t.isValidNode(right) {
			// Current node has a valid right child, so repeat the search from the right child.
			idx = right
		} else {
			// Current node does NOT have a right child, so it must be the rightmost child.
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

func (t *TokenTree) deleteNodeAtIdx(idx int) int {
	if right := rightChildIdx(idx); t.isValidNode(right) {
		replaceIdx := t.leftmostChildInSubtree(right)
		t.nodes[idx] = t.nodes[replaceIdx]
		t.deleteNodeAtIdx(replaceIdx)
		return idx
	} else if left := leftChildIdx(idx); t.isValidNode(left) {
		replaceIdx := t.rightmostChildInSubtree(left)
		t.nodes[idx] = t.nodes[replaceIdx]
		t.deleteNodeAtIdx(replaceIdx)
		return t.nextNodeIdx(idx)
	} else {
		nextIdx := t.nextNodeIdx(idx)
		t.nodes[idx] = treeNode{}
		t.recalculateMaxLookaheadFromIdxToRoot(parentIdx(idx))
		return nextIdx
	}
}

func (t *TokenTree) recalculateMaxLookaheadFromIdxToRoot(idx int) {
	path := make([]int, 0, nodeDepth(idx)+1)
	for idx >= 0 {
		if t.isValidNode(idx) {
			path = append(path, idx)
		}
		idx = parentIdx(idx)
	}

	// Propagate lazy edits from parent to child.
	for i := len(path) - 1; i >= 0; i-- {
		idx := path[i]
		t.propagateLazyEdits(idx)
	}

	// Propagate maxLookahead from child to parent.
	for _, idx := range path {
		node := &t.nodes[idx]
		maxLookaheadPos := node.lookaheadPos()

		if left := leftChildIdx(idx); t.isValidNode(left) {
			t.propagateLazyEdits(left)
			leftMaxLookahead := t.nodes[left].maxLookaheadPos
			if leftMaxLookahead > maxLookaheadPos {
				node.maxLookaheadPos = leftMaxLookahead
			}
		}

		if right := rightChildIdx(idx); t.isValidNode(right) {
			t.propagateLazyEdits(right)
			rightMaxLookahead := t.nodes[right].maxLookaheadPos
			if rightMaxLookahead > maxLookaheadPos {
				node.maxLookaheadPos = rightMaxLookahead
			}
		}
	}
}

// TokenIter iterates over tokens.
// Iterator operations are NOT thread-safe because they can mutate the tree (applying lazy edits).
type TokenIter struct {
	tree    *TokenTree
	nodeIdx int
}

// Get retrieves the current token, if it exists.
func (iter *TokenIter) Get(tok *Token) bool {
	if !iter.tree.isValidNode(iter.nodeIdx) {
		return false
	}

	*tok = iter.tree.nodes[iter.nodeIdx].token
	return true
}

// Advance moves the iterator to the next token.
// If there are no more tokens, this is a no-op.
func (iter *TokenIter) Advance() {
	if iter.tree.isValidNode(iter.nodeIdx) {
		iter.nodeIdx = iter.tree.nextNodeIdx(iter.nodeIdx)
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

// delete removes the token at the iterator's current position and advances to the next token.
// If the iterator is already exhausted, this is a no-op.
func (iter *TokenIter) deleteCurrent() {
	if iter.tree.isValidNode(iter.nodeIdx) {
		iter.nodeIdx = iter.tree.deleteNodeAtIdx(iter.nodeIdx)
	}
}

// deleteToPos deletes tokens starting before the given position.
func (iter *TokenIter) deleteToPos(pos uint64) {
	var tok Token
	for iter.Get(&tok) {
		if tok.StartPos >= pos {
			return
		}
		iter.deleteCurrent()
	}
}

// deleteRemaining deletes all remaining tokens from the iterator.
func (iter *TokenIter) deleteRemaining() {
	for iter.tree.isValidNode(iter.nodeIdx) {
		iter.deleteCurrent()
	}
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

func nodeDepth(idx int) int {
	d := 0
	for idx > 0 {
		idx = idx >> 1
		d++
	}
	return d
}

type treeNode struct {
	initialized     bool
	token           Token
	lazyEdits       []Edit
	maxLookaheadPos uint64
}

func newTreeNodeForToken(token Token) treeNode {
	return treeNode{
		initialized:     true,
		token:           token,
		maxLookaheadPos: token.LookaheadPos,
	}
}

func (tn *treeNode) startPos() uint64 {
	return tn.token.StartPos
}

func (tn *treeNode) endPos() uint64 {
	return tn.token.EndPos
}

func (tn *treeNode) lookaheadPos() uint64 {
	return tn.token.LookaheadPos
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
		tn.maxLookaheadPos = edit.applyToPosition(tn.maxLookaheadPos)
	}
	return edits
}

func (tn *treeNode) applyEdit(edit Edit) {
	// It is possible for deletions to produce tokens with zero length (start pos == end pos).
	tn.token.StartPos = edit.applyToPosition(tn.token.StartPos)
	tn.token.EndPos = edit.applyToPosition(tn.token.EndPos)
	tn.token.LookaheadPos = edit.applyToPosition(tn.token.LookaheadPos)
}
