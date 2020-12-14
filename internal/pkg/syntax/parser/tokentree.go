package parser

// TokenTree represents a collection of tokens.
// It supports efficient lookups by position and "shifting" token positions to account for insertions/deletions.
type TokenTree struct {
	root *innerNode
}

// NewTokenTree constructs a token tree from a set of tokens.
// The tokens must be sorted ascending by start position and cover the entire text.
// Each token's length must be greater than zero.
func NewTokenTree(tokens []Token) *TokenTree {
	leaves := leavesFromTokens(tokens)
	root := buildTreeFromLeaves(leaves)
	return &TokenTree{root}
}

func leavesFromTokens(tokens []Token) []node {
	if len(tokens) == 0 {
		return nil
	}

	var startPos uint64
	leaves := make([]node, 0, len(tokens))
	currentLeaf := &leafNode{}
	for _, tok := range tokens {
		if tok.StartPos != startPos {
			panic("Tokens must be contiguous")
		}
		startPos = tok.EndPos

		tokLength := tok.length()
		if tokLength == 0 {
			panic("Tokens must have non-zero length")
		}

		i := currentLeaf.numEntries
		currentLeaf.entries[i] = entry{
			length:          tokLength,
			lookaheadLength: tok.lookaheadLength(),
		}
		currentLeaf.tokenRoles[i] = tok.Role
		currentLeaf.numEntries++

		if currentLeaf.numEntries == maxEntriesPerLeafNode {
			nextLeaf := &leafNode{}
			currentLeaf.next = nextLeaf
			nextLeaf.prev = currentLeaf
			leaves = append(leaves, currentLeaf)
			currentLeaf = nextLeaf
		}
	}

	if currentLeaf.numEntries > 0 {
		leaves = append(leaves, currentLeaf)
	}

	return leaves
}

func buildTreeFromLeaves(leaves []node) *innerNode {
	if len(leaves) == 0 {
		return nil
	}

	children := leaves
	parents := make([]*innerNode, 0, len(children)/maxEntriesPerInnerNode+1)
	for {
		current := &innerNode{}

		for _, child := range children {
			i := current.numEntries
			current.entries[i] = child.entry()
			current.children[i] = child
			current.numEntries++

			if current.numEntries == maxEntriesPerInnerNode {
				next := &innerNode{}
				parents = append(parents, current)
				current = next
			}
		}

		if current.numEntries > 0 {
			parents = append(parents, current)
		}

		if len(parents) == 1 {
			return parents[0]
		}

		children = children[:0]
		for _, parent := range parents {
			children = append(children, parent)
		}
		parents = parents[:0]
	}
}

// IterFromPosition returns a token iterator from the token intersecting a position.
func (t *TokenTree) IterFromPosition(pos uint64) *TokenIter {
	if t == nil || t.root == nil {
		return nil
	}
	return t.root.iter(0, pos, func(e entry) uint64 {
		return e.length
	})
}

// iterFromFirstAffected returns a token iterator from the first token that could be affected by an edit.
// A token could be affected by an edit if the edit occurred in the half-open interval from the token's start position and its lookahead position.
func (t *TokenTree) iterFromFirstAffected(editPos uint64) *TokenIter {
	if t == nil || t.root == nil {
		return nil
	}
	return t.root.iter(0, editPos, func(e entry) uint64 {
		return e.lookaheadLength + 1
	})
}

// textLen returns the total number of characters represented by tokens in the tree.
func (t *TokenTree) textLen() uint64 {
	t.ensureRootInitialized()
	return t.root.entry().length
}

// insertToken inserts a new token into the tree, shifting the positions of subsequent tokens.
func (t *TokenTree) insertToken(token Token) {
	t.ensureRootInitialized()
	splitNode := t.root.insert(token.StartPos, token)
	if splitNode != nil {
		newRoot := &innerNode{numEntries: 2}
		newRoot.entries[0] = t.root.entry()
		newRoot.children[0] = t.root
		newRoot.entries[1] = splitNode.entry()
		newRoot.children[1] = splitNode
		t.root = newRoot
	}
}

// extendTokenIntersectingPos increases the length of the token intersecting the specified position, shifting the positions of subsequent tokens.
func (t *TokenTree) extendTokenIntersectingPos(pos uint64, extendLen uint64) {
	t.ensureRootInitialized()
	t.root.extendTokenIntersectingPos(pos, extendLen)
}

// deleteRange removes character positions intersecting the range by truncating and removing tokens.
func (t *TokenTree) deleteRange(startPos uint64, numDeleted uint64) {
	if t.root == nil {
		return
	}
	t.root.deleteRange(startPos, numDeleted)
}

func (t *TokenTree) ensureRootInitialized() {
	if t.root == nil || t.root.numEntries == 0 {
		child := &leafNode{}
		t.root = &innerNode{numEntries: 1}
		t.root.entries[0] = child.entry()
		t.root.children[0] = child
	}
}

const maxEntriesPerInnerNode = 42
const maxEntriesPerLeafNode = 50

type node interface {
	// entry returns an entry for the subtree rooted at this node.
	entry() entry

	// iter returns an iterator at the specified position.
	iter(startPos uint64, relativePos uint64, entryLenFunc func(entry) uint64) *TokenIter

	// extendTokenIntersectingPos increases the length of the token at the specified position.
	extendTokenIntersectingPos(relativePos uint64, extendLen uint64)

	// insert inserts a new token into the tree at the specified position.
	insert(relativePos uint64, token Token) node

	// deleteRange deletes character positions intersecting the range by truncating and removing tokens.
	deleteRange(relativePos uint64, remainingLen uint64) uint64

	// leftAndRightLeaves returns the leftmost and rightmost leaves in the subtree rooted at the node.
	leftAndRightLeaves() (*leafNode, *leafNode)
}

type entry struct {
	length          uint64
	lookaheadLength uint64
}

func sumEntries(entries []entry) entry {
	result := entry{}
	for _, e := range entries {
		result.length += e.length
		result.lookaheadLength = maxUint64(result.lookaheadLength, result.length+e.lookaheadLength)
	}
	return result
}

type innerNode struct {
	numEntries int
	entries    [maxEntriesPerInnerNode]entry
	children   [maxEntriesPerLeafNode]node
}

func (in *innerNode) entry() entry {
	return sumEntries(in.entries[0:in.numEntries])
}

func (in *innerNode) iter(startPos uint64, relativePos uint64, entryLenFunc func(entry) uint64) *TokenIter {
	for i := 0; i < in.numEntries; i++ {
		entry := in.entries[i]
		if entryLenFunc(entry) > relativePos {
			return in.children[i].iter(startPos, relativePos, entryLenFunc)
		}
		startPos += entry.length
		relativePos = subtractNoUnderflowUint64(relativePos, entry.length)
	}
	return nil
}

func (in *innerNode) extendTokenIntersectingPos(relativePos uint64, extendLen uint64) {
	for i := 0; i < in.numEntries; i++ {
		entryLen := in.entries[i].length
		if i == in.numEntries-1 || entryLen > relativePos {
			child := in.children[i]
			child.extendTokenIntersectingPos(relativePos, extendLen)
			in.entries[i] = child.entry()
			return
		}
		relativePos -= entryLen
	}
}

func (in *innerNode) insert(relativePos uint64, token Token) node {
	for i := 0; i < in.numEntries-1; i++ {
		entryLen := in.entries[i].length
		if entryLen > relativePos {
			return in.insertAtIdx(i, relativePos, token)
		}
		relativePos -= entryLen
	}
	return in.insertAtIdx(in.numEntries-1, relativePos, token)
}

func (in *innerNode) insertAtIdx(idx int, relativePos uint64, token Token) node {
	child := in.children[idx]
	splitChild := child.insert(relativePos, token)
	in.entries[idx] = child.entry()
	if splitChild == nil {
		return nil
	}

	splitIdx := idx + 1
	if in.numEntries < maxEntriesPerInnerNode {
		in.insertChildAtIdx(splitIdx, splitChild)
		return nil
	}

	var splitNode *innerNode
	if splitIdx == in.numEntries {
		// Fast path with less fragmentation for sequential inserts (common case).
		splitNode = &innerNode{}
	} else {
		// Slow path for non-sequential inserts.
		splitNode = in.splitEvenly()
	}

	if splitIdx < in.numEntries {
		in.insertChildAtIdx(splitIdx, splitChild)
	} else {
		splitNode.insertChildAtIdx(splitIdx-in.numEntries, splitChild)
	}
	return splitNode
}

func (in *innerNode) insertChildAtIdx(idx int, child node) {
	for i := in.numEntries; i > idx; i-- {
		in.entries[i] = in.entries[i-1]
		in.children[i] = in.children[i-1]
	}

	in.entries[idx] = child.entry()
	in.children[idx] = child
	in.numEntries++
}

func (in *innerNode) splitEvenly() *innerNode {
	splitIdx := in.numEntries / 2
	splitNode := &innerNode{numEntries: in.numEntries - splitIdx}
	for i := 0; i < splitNode.numEntries; i++ {
		splitNode.entries[i] = in.entries[splitIdx+i]
		splitNode.children[i] = in.children[splitIdx+i]
	}
	in.numEntries = splitIdx
	return splitNode
}

func (in *innerNode) deleteRange(relativePos uint64, remainingLen uint64) uint64 {
	initialRemainingLen := remainingLen
	childrenToDelete := make([]int, 0, in.numEntries)
	for i := 0; i < in.numEntries && remainingLen > 0; i++ {
		entryLen := in.entries[i].length
		numToDelete := minUint64(subtractNoUnderflowUint64(entryLen, relativePos), remainingLen)
		if numToDelete == 0 {
			// This occurs when the start position of the deleted range starts after the subtree rooted at this child.
			// Skip this token and keep looking.
			relativePos = subtractNoUnderflowUint64(relativePos, entryLen)
			continue
		}
		if numToDelete == entryLen {
			// If the subtree rooted at this child is entirely within the deleted range, mark the child for removal.
			childrenToDelete = append(childrenToDelete, i)
			relativePos = subtractNoUnderflowUint64(relativePos, entryLen)
			remainingLen -= entryLen
			continue
		}

		// The subtree rooted at this child partially intersects the deleted range, so recursively delete from the subtree.
		numDeleted := in.children[i].deleteRange(relativePos, remainingLen)
		in.entries[i].length -= numDeleted
		in.entries[i].lookaheadLength -= numDeleted
		remainingLen = subtractNoUnderflowUint64(remainingLen, numDeleted)
		relativePos = subtractNoUnderflowUint64(relativePos, entryLen)
	}

	// Remove children we marked earlier.
	// We need to both remove the child from the inner node and unlink the child's leaves from the linked list of leaves.
	for i := len(childrenToDelete) - 1; i >= 0; i-- {
		deleteIdx := childrenToDelete[i]
		in.unlinkLeaves(deleteIdx)
		for j := deleteIdx; j < in.numEntries-1; j++ {
			in.entries[j] = in.entries[j+1]
			in.children[j] = in.children[j+1]
		}
		in.numEntries--
	}

	return initialRemainingLen - remainingLen
}

func (in *innerNode) unlinkLeaves(childIdx int) {
	leftLeaf, rightLeaf := in.children[childIdx].leftAndRightLeaves()

	if leftLeaf.prev != nil {
		leftLeaf.prev.next = rightLeaf.next
	}

	if rightLeaf.next != nil {
		rightLeaf.next.prev = leftLeaf.prev
	}
}

func (in *innerNode) leftAndRightLeaves() (*leafNode, *leafNode) {
	leftChild := in.children[0]
	leftLeaf, _ := leftChild.leftAndRightLeaves()

	rightChild := in.children[in.numEntries-1]
	_, rightLeaf := rightChild.leftAndRightLeaves()

	return leftLeaf, rightLeaf
}

type leafNode struct {
	numEntries int
	entries    [maxEntriesPerLeafNode]entry
	tokenRoles [maxEntriesPerLeafNode]TokenRole
	prev       *leafNode
	next       *leafNode
}

func (ln *leafNode) entry() entry {
	return sumEntries(ln.entries[0:ln.numEntries])
}

func (ln *leafNode) iter(startPos uint64, relativePos uint64, entryLenFunc func(entry) uint64) *TokenIter {
	for i := 0; i < ln.numEntries; i++ {
		entry := ln.entries[i]
		if entryLenFunc(entry) > relativePos {
			return &TokenIter{
				leaf:        ln,
				entryOffset: i,
				startPos:    startPos,
			}
		}
		startPos += entry.length
		relativePos = subtractNoUnderflowUint64(relativePos, entry.length)
	}
	return nil
}

func (ln *leafNode) extendTokenIntersectingPos(relativePos uint64, extendLen uint64) {
	for i := 0; i < ln.numEntries; i++ {
		entry := &ln.entries[i]
		if entry.length > relativePos {
			entry.length += extendLen
			entry.lookaheadLength += extendLen
			return
		}
		relativePos = subtractNoUnderflowUint64(relativePos, entry.length)
	}
	panic("Could not find token to extend at position")
}

func (ln *leafNode) insert(relativePos uint64, token Token) node {
	for i := 0; i < ln.numEntries; i++ {
		entryLen := ln.entries[i].length
		if entryLen > relativePos {
			return ln.insertAtIdx(i, token)
		}
		relativePos = subtractNoUnderflowUint64(relativePos, entryLen)
	}
	return ln.insertAtIdx(ln.numEntries, token)
}

func (ln *leafNode) insertAtIdx(idx int, token Token) node {
	if ln.numEntries < maxEntriesPerLeafNode {
		ln.insertAtIdxNoSplit(idx, token)
		return nil
	}

	var splitNode *leafNode
	if idx == ln.numEntries {
		// Fast path with less fragmentation for sequential inserts (common case).
		splitNode = ln.splitAppendEmptyNode()
	} else {
		// Slow path for non-sequential inserts.
		splitNode = ln.splitEvenly()
	}

	if idx < ln.numEntries {
		ln.insertAtIdxNoSplit(idx, token)
	} else {
		splitNode.insertAtIdxNoSplit(idx-ln.numEntries, token)
	}
	return splitNode
}

func (ln *leafNode) insertAtIdxNoSplit(idx int, token Token) {
	for i := ln.numEntries; i > idx; i-- {
		ln.entries[i] = ln.entries[i-1]
		ln.tokenRoles[i] = ln.tokenRoles[i-1]
	}

	ln.entries[idx] = entry{
		length:          token.length(),
		lookaheadLength: token.lookaheadLength(),
	}
	ln.tokenRoles[idx] = token.Role
	ln.numEntries++
}

func (ln *leafNode) splitAppendEmptyNode() *leafNode {
	splitNode := &leafNode{}
	ln.next = splitNode
	splitNode.prev = ln
	return splitNode
}

func (ln *leafNode) splitEvenly() *leafNode {
	splitIdx := ln.numEntries / 2
	splitNode := &leafNode{numEntries: ln.numEntries - splitIdx}
	for i := 0; i < splitNode.numEntries; i++ {
		splitNode.entries[i] = ln.entries[splitIdx+i]
		splitNode.tokenRoles[i] = ln.tokenRoles[splitIdx+i]
	}
	splitNode.next = ln.next
	splitNode.prev = ln
	ln.next = splitNode
	ln.numEntries = splitIdx
	return splitNode
}

func (ln *leafNode) deleteRange(relativePos uint64, remainingLen uint64) uint64 {
	initialRemainingLen := remainingLen
	tokensToDelete := make([]int, 0, ln.numEntries)
	for i := 0; i < ln.numEntries && remainingLen > 0; i++ {
		entryLen := ln.entries[i].length
		numToDelete := minUint64(subtractNoUnderflowUint64(entryLen, relativePos), remainingLen)
		relativePos = subtractNoUnderflowUint64(relativePos, entryLen)
		if numToDelete == 0 {
			// The start of the deletion range occurs after this token, so skip it and keep looking.
			continue
		}
		if numToDelete == entryLen {
			// The token is entirely within the deletion range, so mark it for removal.
			tokensToDelete = append(tokensToDelete, i)
			remainingLen = subtractNoUnderflowUint64(remainingLen, entryLen)
			continue
		}

		// The token partially overlaps the deletion range, so truncate it.
		ln.entries[i].length -= numToDelete
		ln.entries[i].lookaheadLength -= numToDelete
		remainingLen -= numToDelete
	}

	// Remove the tokens we marked earlier.
	for i := len(tokensToDelete) - 1; i >= 0; i-- {
		deleteIdx := tokensToDelete[i]
		for j := deleteIdx; j < ln.numEntries-1; j++ {
			ln.entries[j] = ln.entries[j+1]
			ln.tokenRoles[j] = ln.tokenRoles[j+1]
		}
		ln.numEntries--
	}

	return initialRemainingLen - remainingLen
}

func (ln *leafNode) leftAndRightLeaves() (*leafNode, *leafNode) {
	return ln, ln
}

// TokenIter iterates over tokens.
// Iterator operations are NOT thread-safe because they can mutate the tree (applying lazy edits).
type TokenIter struct {
	leaf        *leafNode
	entryOffset int
	startPos    uint64
}

// Get retrieves the current token, if it exists.
func (iter *TokenIter) Get(tok *Token) bool {
	if iter == nil || iter.leaf == nil {
		return false
	}

	entry := iter.leaf.entries[iter.entryOffset]
	role := iter.leaf.tokenRoles[iter.entryOffset]
	*tok = Token{
		StartPos:     iter.startPos,
		EndPos:       iter.startPos + entry.length,
		LookaheadPos: iter.startPos + entry.lookaheadLength,
		Role:         role,
	}
	return true
}

// Advance moves the iterator to the next token.
// If there are no more tokens, this is a no-op.
func (iter *TokenIter) Advance() {
	if iter == nil || iter.leaf == nil {
		return
	}

	entry := iter.leaf.entries[iter.entryOffset]
	iter.startPos += entry.length
	iter.entryOffset++
	for iter.leaf != nil && iter.entryOffset >= iter.leaf.numEntries {
		iter.leaf = iter.leaf.next
		iter.entryOffset = 0
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
