package text

import (
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

// text.Tree is a data structure for representing UTF-8 text.
// It supports efficient insertions, deletions, and lookup by character offset and line number.
// It is inspired by two papers:
// Boehm, H. J., Atkinson, R., & Plass, M. (1995). Ropes: an alternative to strings. Software: Practice and Experience, 25(12), 1315-1330.
// Rao, J., & Ross, K. A. (2000, May). Making B+-trees cache conscious in main memory. In Proceedings of the 2000 ACM SIGMOD international conference on Management of data (pp. 475-486).
// Like a rope, the tree maintains character counts at each level to efficiently locate a character at a given offset.
// To use the CPU cache efficiently, all children of a node are pre-allocated in a group (what the Rao & Ross paper calls a "full" cache-sensitive B+ tree),
// and the parent uses offsets within the node group to identify child nodes.
// All nodes are carefully designed to fit as much data as possible within a 64-byte cache line.
type Tree struct {
	root *innerNode
}

// NewTree returns a tree representing an empty string.
func NewTree() *Tree {
	root := &innerNode{numKeys: 1}
	root.child = &leafNodeGroup{numNodes: 1}
	return &Tree{root}
}

// NewTreeFromReader creates a new Tree from a reader that produces UTF-8 text.
// This is more efficient than inserting the bytes into an empty tree.
// Returns an error if the bytes are invalid UTF-8.
func NewTreeFromReader(r io.Reader) (*Tree, error) {
	leafGroups, err := bulkLoadIntoLeaves(r)
	if err != nil {
		return nil, err
	}
	root := buildTreeFromLeaves(leafGroups)
	return &Tree{root}, nil
}

// NewTreeFromString creates a new Tree from a UTF-8 string.
func NewTreeFromString(s string) (*Tree, error) {
	reader := strings.NewReader(s)
	return NewTreeFromReader(reader)
}

func bulkLoadIntoLeaves(r io.Reader) ([]nodeGroup, error) {
	v := NewValidator()
	leafGroups := make([]nodeGroup, 0, 1)
	currentGroup := &leafNodeGroup{numNodes: 1}
	currentNode := &currentGroup.nodes[0]
	leafGroups = append(leafGroups, currentGroup)

	var buf [1024]byte
	for {
		n, err := r.Read(buf[:])
		if err != nil && err != io.EOF {
			return nil, err
		}

		if n == 0 {
			break
		}

		if !v.ValidateBytes(buf[:n]) {
			return nil, errors.New("invalid UTF-8")
		}

		for i := 0; i < n; i++ {
			charWidth := utf8CharWidth[buf[i]] // zero for continuation bytes
			if currentNode.numBytes+charWidth >= maxBytesPerLeaf {
				if currentGroup.numNodes < maxNodesPerGroup {
					currentNode = &currentGroup.nodes[currentGroup.numNodes]
					currentGroup.numNodes++
				} else {
					newGroup := &leafNodeGroup{numNodes: 1}
					leafGroups = append(leafGroups, newGroup)
					newGroup.prev = currentGroup
					currentGroup.next = newGroup
					currentGroup = newGroup
					currentNode = &currentGroup.nodes[0]
				}
			}

			currentNode.textBytes[currentNode.numBytes] = buf[i]
			currentNode.numBytes++
		}
	}

	if !v.ValidateEnd() {
		return nil, errors.New("invalid UTF-8")
	}

	return leafGroups, nil
}

func buildTreeFromLeaves(leafGroups []nodeGroup) *innerNode {
	childGroups := leafGroups

	for {
		parentGroups := make([]nodeGroup, 0, len(childGroups)/maxNodesPerGroup+1)
		currentGroup := &innerNodeGroup{}
		parentGroups = append(parentGroups, currentGroup)

		for _, cg := range childGroups {
			if currentGroup.numNodes == maxNodesPerGroup {
				newGroup := &innerNodeGroup{}
				parentGroups = append(parentGroups, newGroup)
				currentGroup = newGroup
			}

			innerNode := &currentGroup.nodes[currentGroup.numNodes]
			innerNode.child = cg
			innerNode.recalculateChildKeys()
			currentGroup.numNodes++
		}

		if len(parentGroups) == 1 {
			root := innerNode{child: parentGroups[0]}
			root.recalculateChildKeys()
			return &root
		}

		childGroups = parentGroups
	}
}

// InsertAtPosition inserts a UTF-8 character at the specified position (0-indexed).
// If charPos is past the end of the text, it will be appended at the end.
// Returns an error if c is not a valid UTF-8 character.
func (t *Tree) InsertAtPosition(charPos uint64, c rune) error {
	invalidateKeys, splitNode, err := t.root.insertAtPosition(charPos, c)
	if err != nil {
		return err
	}

	if invalidateKeys {
		t.root.recalculateChildKeys()
	}

	if splitNode != nil {
		newGroup := innerNodeGroup{numNodes: 2}
		newGroup.nodes[0] = *t.root
		newGroup.nodes[1] = *splitNode

		t.root = &innerNode{child: &newGroup}
		t.root.recalculateChildKeys()
	}

	return nil
}

// DeleteAtPosition removes the UTF-8 character at the specified position (0-indexed).
// If charPos is past the end of the text, this has no effect.
func (t *Tree) DeleteAtPosition(charPos uint64) {
	t.root.deleteAtPosition(charPos)
}

// CursorAtPosition returns a cursor starting at the UTF-8 character at the specified position (0-indexed).
// If the position is past the end of the text, the returned cursor will read zero bytes.
func (t *Tree) CursorAtPosition(charPos uint64) *Cursor {
	return t.root.cursorAtPosition(charPos)
}

// CursorAtLine returns a cursor starting at the first character at the specified line (0-indexed).
// For line zero, this is the first character in the tree; for subsequent lines, this is the first
// character after the newline character.
// If the line number is greater than the maximum line number, the returned cursor will read zero bytes.
func (t *Tree) CursorAtLine(lineNum uint64) *Cursor {
	if lineNum == 0 {
		// Special case the first line, since it's the only line that doesn't immediately follow a newline character.
		return t.root.cursorAtPosition(0)
	}

	return t.root.cursorAfterNewline(lineNum - 1)
}

// text.Cursor reads UTF-8 bytes from a text.Tree.
// It implements io.Reader.
// text.Tree is NOT thread-safe, so reading from a tree while modifying it is undefined behavior!
type Cursor struct {
	group          *leafNodeGroup
	nodeIdx        uint64
	textByteOffset uint64
}

func (c *Cursor) Read(b []byte) (int, error) {
	i := 0
	for {
		if i == len(b) {
			return i, nil
		}

		if c.group == nil {
			return i, io.EOF
		}

		node := &c.group.nodes[c.nodeIdx]
		bytesWritten := copy(b[i:], node.textBytes[c.textByteOffset:node.numBytes])
		c.textByteOffset += uint64(bytesWritten)
		i += bytesWritten

		if c.textByteOffset == uint64(node.numBytes) {
			c.nodeIdx++
			c.textByteOffset = 0
		}

		if c.nodeIdx == c.group.numNodes {
			c.group = c.group.next
			c.nodeIdx = 0
			c.textByteOffset = 0
		}
	}

	return 0, nil
}

const maxKeysPerNode = 64
const maxNodesPerGroup = maxKeysPerNode
const maxBytesPerLeaf = 63

// nodeGroup is either an inner node group or a leaf node group.
type nodeGroup interface {
	keys() []indexKey
	insertAtPosition(nodeIdx uint64, charPos uint64, c rune) (invalidateKeys bool, splitNodeGroup nodeGroup, err error)
	deleteAtPosition(nodeIdx uint64, charPos uint64) (didDelete, wasNewline bool)
	cursorAtPosition(nodeIdx uint64, charPos uint64) *Cursor
	cursorAfterNewline(nodeIdx uint64, newlinePos uint64) *Cursor
}

// indexKey is used to navigate from an inner node to the child node containing a particular line or character offset.
type indexKey struct {

	// Number of UTF-8 characters in a subtree.
	numChars uint64

	// Number of newline characters in a subtree.
	numNewlines uint64
}

// innerNodeGroup is a group of inner nodes referenced by a parent inner node.
type innerNodeGroup struct {
	numNodes uint64
	nodes    [maxNodesPerGroup]innerNode
}

func (g *innerNodeGroup) keys() []indexKey {
	keys := make([]indexKey, g.numNodes)
	for i := uint64(0); i < g.numNodes; i++ {
		keys[i] = g.nodes[i].key()
	}
	return keys
}

func (g *innerNodeGroup) insertAtPosition(nodeIdx uint64, charPos uint64, c rune) (invalidateKeys bool, splitNodeGroup nodeGroup, err error) {
	invalidateKeys, splitNode, err := g.nodes[nodeIdx].insertAtPosition(charPos, c)
	if err != nil {
		return false, nil, err
	}

	if splitNode == nil {
		return false, nil, nil
	}

	splitIdx := nodeIdx + 1
	if g.numNodes < maxNodesPerGroup {
		g.insertNode(splitIdx, splitNode)
		return true, nil, nil
	}

	splitGroup := g.split()
	if splitIdx < g.numNodes {
		g.insertNode(splitIdx, splitNode)
	} else {
		splitGroup.insertNode(splitIdx-g.numNodes, splitNode)
	}

	return true, splitGroup, nil
}

func (g *innerNodeGroup) insertNode(nodeIdx uint64, node *innerNode) {
	for i := int(g.numNodes) - 1; i > int(nodeIdx); i-- {
		g.nodes[i] = g.nodes[i-1]
	}
	g.nodes[nodeIdx] = *node
	g.numNodes++
}

func (g *innerNodeGroup) split() *innerNodeGroup {
	mid := g.numNodes / 2
	splitGroup := innerNodeGroup{numNodes: g.numNodes - mid}
	for i := uint64(0); i < splitGroup.numNodes; i++ {
		splitGroup.nodes[i] = g.nodes[mid+i]
	}
	g.numNodes = mid
	return &splitGroup
}

func (g *innerNodeGroup) deleteAtPosition(nodeIdx uint64, charPos uint64) (didDelete, wasNewline bool) {
	return g.nodes[nodeIdx].deleteAtPosition(charPos)
}

func (g *innerNodeGroup) cursorAtPosition(nodeIdx uint64, charPos uint64) *Cursor {
	return g.nodes[nodeIdx].cursorAtPosition(charPos)
}

func (g *innerNodeGroup) cursorAfterNewline(nodeIdx uint64, newlinePos uint64) *Cursor {
	return g.nodes[nodeIdx].cursorAfterNewline(newlinePos)
}

// innerNode is used to navigate to the leaf node containing a character offset or line number.
//
// +-----------------------------+
// | child | numKeys |  keys[64] |
// +-----------------------------+
//     8        8         1024     = 1032 bytes
//
type innerNode struct {
	child   nodeGroup
	numKeys uint64

	// Each key corresponds to a node in the child group.
	keys [maxKeysPerNode]indexKey
}

func (n *innerNode) key() indexKey {
	nodeKey := indexKey{}
	for i := uint64(0); i < n.numKeys; i++ {
		key := n.keys[i]
		nodeKey.numChars += key.numChars
		nodeKey.numNewlines += key.numNewlines
	}
	return nodeKey
}

func (n *innerNode) recalculateChildKeys() {
	childKeys := n.child.keys()
	for i, key := range childKeys {
		n.keys[i] = key
	}
	n.numKeys = uint64(len(childKeys))
}

func (n *innerNode) insertAtPosition(charPos uint64, c rune) (invalidateKeys bool, splitNode *innerNode, err error) {
	nodeIdx, adjustedCharPos := n.locatePosition(charPos)

	invalidateKeys, splitGroup, err := n.child.insertAtPosition(nodeIdx, adjustedCharPos, c)
	if err != nil {
		return false, nil, err
	}

	if invalidateKeys {
		n.recalculateChildKeys()
	} else {
		key := &n.keys[nodeIdx]
		key.numChars++
		if c == '\n' {
			key.numNewlines++
		}
	}

	if splitGroup == nil {
		return false, nil, nil
	}

	splitNode = &innerNode{child: splitGroup}
	splitNode.recalculateChildKeys()
	return true, splitNode, nil
}

func (n *innerNode) deleteAtPosition(charPos uint64) (didDelete, wasNewline bool) {
	nodeIdx, adjustedCharPos := n.locatePosition(charPos)
	didDelete, wasNewline = n.child.deleteAtPosition(nodeIdx, adjustedCharPos)
	if didDelete {
		n.keys[nodeIdx].numChars--
		if wasNewline {
			n.keys[nodeIdx].numNewlines--
		}
	}
	return
}

func (n *innerNode) cursorAtPosition(charPos uint64) *Cursor {
	nodeIdx, adjustedCharPos := n.locatePosition(charPos)
	return n.child.cursorAtPosition(nodeIdx, adjustedCharPos)
}

func (n *innerNode) cursorAfterNewline(newlinePos uint64) *Cursor {
	c := uint64(0)
	for i := uint64(0); i < n.numKeys-1; i++ {
		nc := n.keys[i].numNewlines
		if newlinePos < c+nc {
			return n.child.cursorAfterNewline(i, newlinePos-c)
		}
		c += nc
	}

	return n.child.cursorAfterNewline(n.numKeys-1, newlinePos-c)
}

func (n *innerNode) locatePosition(charPos uint64) (nodeIdx, adjustedCharPos uint64) {
	c := uint64(0)
	for i := uint64(0); i < n.numKeys; i++ {
		nc := n.keys[i].numChars
		if charPos < c+nc {
			return i, charPos - c
		}
		c += nc
	}
	return n.numKeys - 1, c
}

// leafNodeGroup is a group of leaf nodes referenced by an inner node.
// These form a doubly-linked list so a cursor can scan the text efficiently.
type leafNodeGroup struct {
	prev     *leafNodeGroup
	next     *leafNodeGroup
	numNodes uint64
	nodes    [maxNodesPerGroup]leafNode
}

func (g *leafNodeGroup) keys() []indexKey {
	keys := make([]indexKey, g.numNodes)
	for i := uint64(0); i < g.numNodes; i++ {
		keys[i] = g.nodes[i].key()
	}
	return keys
}

func (g *leafNodeGroup) insertAtPosition(nodeIdx uint64, charPos uint64, c rune) (invalidateKeys bool, splitNodeGroup nodeGroup, err error) {
	splitNode, err := g.nodes[nodeIdx].insertAtPosition(charPos, c)
	if err != nil {
		return false, nil, err
	}

	if splitNode == nil {
		return false, nil, nil
	}

	splitNodeIdx := nodeIdx + 1
	if g.numNodes < maxNodesPerGroup {
		g.insertNode(splitNodeIdx, splitNode)
		return true, nil, nil
	}

	splitGroup := g.split()
	if splitNodeIdx < g.numNodes {
		g.insertNode(splitNodeIdx, splitNode)
	} else {
		splitGroup.insertNode(splitNodeIdx-g.numNodes, splitNode)
	}
	return true, splitGroup, nil
}

func (g *leafNodeGroup) insertNode(nodeIdx uint64, node *leafNode) {
	for i := int(g.numNodes) - 1; i > int(nodeIdx); i-- {
		g.nodes[i] = g.nodes[i-1]
	}
	g.nodes[nodeIdx] = *node
	g.numNodes++
}

func (g *leafNodeGroup) split() *leafNodeGroup {
	mid := g.numNodes / 2
	splitGroup := &leafNodeGroup{numNodes: g.numNodes - mid}
	for i := uint64(0); i < splitGroup.numNodes; i++ {
		splitGroup.nodes[i] = g.nodes[mid+i]
	}
	g.numNodes = mid
	if g.next != nil {
		g.next.prev = splitGroup
		splitGroup.next = g.next
	}
	splitGroup.prev = g
	g.next = splitGroup
	return splitGroup
}

func (g *leafNodeGroup) deleteAtPosition(nodeIdx uint64, charPos uint64) (didDelete, wasNewline bool) {
	// Don't bother rebalancing the tree.  This leaves extra space in the leaves,
	// but that's okay because usually the user will want to insert more text anyway.
	return g.nodes[nodeIdx].deleteAtPosition(charPos)
}

func (g *leafNodeGroup) cursorAtPosition(nodeIdx uint64, charPos uint64) *Cursor {
	textByteOffset := g.nodes[nodeIdx].byteOffsetForPosition(charPos)
	return &Cursor{
		group:          g,
		nodeIdx:        nodeIdx,
		textByteOffset: textByteOffset,
	}
}

func (g *leafNodeGroup) cursorAfterNewline(nodeIdx uint64, newlinePos uint64) *Cursor {
	textByteOffset := g.nodes[nodeIdx].byteOffsetAfterNewline(newlinePos)
	return &Cursor{
		group:          g,
		nodeIdx:        nodeIdx,
		textByteOffset: textByteOffset,
	}
}

// leafNode is a node that stores UTF-8 text as a byte array.
//
// Multi-byte UTF-8 characters are never split between leaf nodes.
//
// +---------------------------------+
// |   numBytes  |   textBytes[63]   |
// +---------------------------------+
//        1               63          = 64 bytes
//
type leafNode struct {
	numBytes  byte
	textBytes [maxBytesPerLeaf]byte
}

func (l *leafNode) key() indexKey {
	key := indexKey{}
	for _, b := range l.textBytes[:l.numBytes] {
		key.numChars += uint64(utf8StartByteIndicator[b])
		if b == '\n' {
			key.numNewlines++
		}
	}
	return key
}

func (l *leafNode) insertAtPosition(charPos uint64, c rune) (*leafNode, error) {
	w := utf8.RuneLen(c)
	if w < 0 {
		return nil, errors.New("invalid UTF-8")
	}

	charWidth := uint64(w)

	if uint64(l.numBytes)+charWidth <= maxBytesPerLeaf {
		l.insertAtPositionNoSplit(charPos, charWidth, c)
		return nil, nil
	}

	splitNode, numCharsRemaining := l.split()
	if charPos < numCharsRemaining {
		l.insertAtPositionNoSplit(charPos, charWidth, c)
	} else {
		splitNode.insertAtPositionNoSplit(charPos-numCharsRemaining, charWidth, c)
	}

	return splitNode, nil
}

func (l *leafNode) insertAtPositionNoSplit(charPos uint64, charWidth uint64, c rune) {
	offset := l.byteOffsetForPosition(charPos)
	l.numBytes += byte(charWidth)
	for i := int(l.numBytes) - 1; i >= int(offset+charWidth); i-- {
		l.textBytes[i] = l.textBytes[i-int(charWidth)]
	}
	utf8.EncodeRune(l.textBytes[offset:], c)
}

func (l *leafNode) split() (*leafNode, uint64) {
	splitIdx, numCharsBeforeSplit := l.splitIdx()
	splitNode := leafNode{numBytes: l.numBytes - splitIdx}
	for i := byte(0); i < splitNode.numBytes; i++ {
		splitNode.textBytes[i] = l.textBytes[i+splitIdx]
	}
	l.numBytes = splitIdx
	return &splitNode, uint64(numCharsBeforeSplit)
}

func (l *leafNode) splitIdx() (splitIdx, numCharsBeforeSplit byte) {
	mid := l.numBytes / 2
	for i := byte(0); i < l.numBytes; i++ {
		b := l.textBytes[i]
		isStartByte := utf8StartByteIndicator[b] > 0
		if i > mid && isStartByte {
			return i, numCharsBeforeSplit
		} else if isStartByte {
			numCharsBeforeSplit++
		}
	}
	return l.numBytes, numCharsBeforeSplit
}

func (l *leafNode) deleteAtPosition(charPos uint64) (didDelete, wasNewline bool) {
	offset := l.byteOffsetForPosition(charPos)
	if offset < uint64(l.numBytes) {
		startByte := l.textBytes[offset]
		charWidth := utf8CharWidth[startByte]
		for i := offset; i < uint64(l.numBytes-charWidth); i++ {
			l.textBytes[i] = l.textBytes[i+uint64(charWidth)]
		}
		l.numBytes -= charWidth
		didDelete = true
		wasNewline = startByte == '\n'
	}
	return
}

func (l *leafNode) byteOffsetForPosition(charPos uint64) uint64 {
	n := uint64(0)
	for i, b := range l.textBytes[:l.numBytes] {
		c := uint64(utf8StartByteIndicator[b])
		if c > 0 && n == charPos {
			return uint64(i)
		}
		n += c
	}
	return uint64(l.numBytes)
}

func (l *leafNode) byteOffsetAfterNewline(newlinePos uint64) uint64 {
	n := uint64(0)
	for i, b := range l.textBytes[:l.numBytes] {
		if b == '\n' {
			if n == newlinePos {
				return uint64(i + 1)
			}
			n++
		}
	}
	return uint64(l.numBytes)
}

// Lookup table for UTF-8 character byte counts.  Set to the byte count of the character for start bytes, zero otherwise.
var utf8CharWidth [256]byte

func init() {
	for b := 0; b < 256; b++ {
		if b>>7 == 0 {
			utf8CharWidth[b] = 1
		} else if b>>5 == 0b110 {
			utf8CharWidth[b] = 2
		} else if b>>4 == 0b1110 {
			utf8CharWidth[b] = 3
		} else if b>>3 == 0b11110 {
			utf8CharWidth[b] = 4
		}
	}
}

// Lookup table for UTF-8 start bytes. Set to 1 for start bytes, zero otherwise.
var utf8StartByteIndicator [256]byte

func init() {
	for b := 0; b < 256; b++ {
		if b>>7 == 0 ||
			b>>5 == 0b110 ||
			b>>4 == 0b1110 ||
			b>>3 == 0b11110 {
			utf8StartByteIndicator[b] = 1
		}
	}
}
