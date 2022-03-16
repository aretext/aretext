package fuzzy

import (
	"math"
	"sort"
)

const (
	// Maximum edit distance in the prefix to match a string in the trie.
	editDistThreshold = 2.0

	// Edit distance costs.
	// Errors in the prefix are more expensive than adding a suffix.
	editDistForDeleteQueryChar   = 1.0
	editDistForDeletePrefixChar  = 1.0
	editDistForReplacePrefixChar = 1.0
	editDistPerSuffixChar        = 0.1
)

// trie is a data structure that fuzzy-matches a query prefix to a set of strings.
// It caches the results of prior queries so it can return results quickly as the user types.
// The algorithm is based on Ji, Shengyue, et al. "Efficient interactive fuzzy keyword search" (2009).
type trie struct {
	nodes           []trieNode              // Trie nodes are identified by index in this slice.
	activeNodeCache map[string][]activeNode // Query prefix -> active nodes
}

// newTrie returns an empty trie.
func newTrie() *trie {
	rootNode := trieNode{
		minRecordId: math.MaxInt,
		maxRecordId: 0,
	}
	return &trie{
		nodes:           []trieNode{rootNode},
		activeNodeCache: make(map[string][]activeNode, 0),
	}
}

// insert adds a new string to the trie and associates it with a record ID.
func (t *trie) insert(s string, recordId int) {
	var currentNodeId int // Start at the root.
	for len(s) > 0 {
		t.nodes[currentNodeId].updateMinAndMaxRecordId(recordId)
		childNodeId, ok := t.findChildForChar(currentNodeId, s[0])
		if !ok {
			// Add a child for this string.
			childNodeId = len(t.nodes)
			t.nodes = append(t.nodes, trieNode{
				recordSlice: s,
				minRecordId: recordId,
				maxRecordId: recordId,
			})
			t.nodes[currentNodeId].childNodeIds.add(childNodeId)
			currentNodeId = childNodeId
			break
		}

		child := &(t.nodes[childNodeId])
		n := sharedPrefixLen(s, child.recordSlice)
		if n == len(child.recordSlice) {
			currentNodeId = childNodeId
		} else {
			// Add a new node representing the non-shared suffix from the child.
			suffixNodeId := len(t.nodes)
			suffixNode := trieNode{
				recordSlice:  child.recordSlice[n:len(child.recordSlice)],
				childNodeIds: child.childNodeIds,
				recordIds:    child.recordIds,
				minRecordId:  child.minRecordId,
				maxRecordId:  child.maxRecordId,
			}
			t.nodes = append(t.nodes, suffixNode)

			// Truncate the child so it represents the shared suffix.
			child.recordSlice = child.recordSlice[0:n]
			child.childNodeIds = []int{suffixNodeId}
			child.recordIds = nil

			// Continue from the truncated child.
			currentNodeId = childNodeId
		}

		// Consume the shared prefix.
		s = s[n:]
	}

	t.nodes[currentNodeId].recordIds.add(recordId)
}

func (t *trie) findChildForChar(nodeId int, c byte) (int, bool) {
	for _, childNodeId := range t.nodes[nodeId].childNodeIds {
		child := t.nodes[childNodeId]
		if child.recordSlice[0] == c {
			return childNodeId, true
		}
	}
	return 0, false
}

// recordIdsForPrefix finds records in the trie within editDistThreshold of a keyword prefix.
// If prevRecordIds is non-nil, it includes only records that are in prevRecordIds.
func (t *trie) recordIdsForPrefix(prefix string, prevRecordIds *recordIdSet) *recordIdSet {
	recordIds := newRecordIdSet()
	visitedNodeIds := newIntSet()
	activeNodes := t.activeNodesForPrefix(prefix)

	var currentNodeId int
	stack := make([]int, 0, len(activeNodes))
	for _, an := range activeNodes {
		stack = append(stack, an.nodeId)
	}
	for len(stack) > 0 {
		currentNodeId, stack = stack[len(stack)-1], stack[0:len(stack)-1]
		if visitedNodeIds.contains(currentNodeId) {
			continue
		}

		node := t.nodes[currentNodeId]
		for _, recordId := range node.recordIds {
			if prevRecordIds != nil && !prevRecordIds.contains(recordId) {
				// Filter by prevRecordIds.
				continue
			}
			recordIds.add(recordId)
		}

		visitedNodeIds.add(currentNodeId)
		for _, childNodeId := range node.childNodeIds {
			child := t.nodes[childNodeId]
			if prevRecordIds != nil && (prevRecordIds.max < child.minRecordId || prevRecordIds.min > child.maxRecordId) {
				// If the child doesn't have any records that match the filter, skip it.
				// This is a performance optimization, not required for correctness.
				continue
			}
			stack = append(stack, childNodeId)
		}
	}
	return recordIds
}

// activeNodesForPrefix calculates the active nodes in the trie
// within the edit distance threshold of a query prefix.
// When possible, it reuses cached results from prior queries.
func (t *trie) activeNodesForPrefix(prefix string) []activeNode {
	if ans, ok := t.activeNodeCache[prefix]; ok {
		// Re-use active nodes calculated from an earlier query.
		return ans
	}

	ans := make([]activeNode, 0)

	// Helper function to add all nodes in a subtree rooted at nodeId that are within the edit distance threshold.
	addDescendantsWithinEditDistThreshold := func(root activeNode) {
		var current activeNode
		stack := []activeNode{root}
		for len(stack) > 0 {
			current, stack = stack[len(stack)-1], stack[0:len(stack)-1]
			if current.editDist <= editDistThreshold {
				ans = append(ans, current)
				t.expandActiveNode(current, func(nodeId int, recordSliceLen int) {
					stack = append(stack, activeNode{
						nodeId:         nodeId,
						recordSliceLen: recordSliceLen,
						editDist:       current.editDist + editDistForDeletePrefixChar,
					})
				})
			}
		}
	}

	// Recursively build the active node set from an empty string,
	// adding one character from the prefix at a time.
	if len(prefix) == 0 {
		// Base case: empty prefix.
		addDescendantsWithinEditDistThreshold(activeNode{
			nodeId:         0, // root
			recordSliceLen: 0,
			editDist:       0,
		})
	} else {
		// Recursive case: expand the active nodes by one character.
		prevActiveNodes := t.activeNodesForPrefix(prefix[0 : len(prefix)-1])
		for _, an := range prevActiveNodes {
			// Delete the next character from the query, increasing the edit distance.
			deletionEditDist := an.editDist + editDistForDeleteQueryChar
			if deletionEditDist <= editDistThreshold {
				ans = append(ans, activeNode{
					nodeId:         an.nodeId,
					recordSliceLen: an.recordSliceLen,
					editDist:       deletionEditDist,
				})
			}

			t.expandActiveNode(an, func(nodeId int, recordSliceLen int) {
				nextChar := t.nodes[nodeId].recordSlice[recordSliceLen-1]
				if prefix[len(prefix)-1] == nextChar {
					// The next character matches the prefix, so the child is an
					// active node with the same edit distance as the parent.
					// We also need to include all descendants of the child within
					// the edit distance threshold.
					addDescendantsWithinEditDistThreshold(activeNode{
						nodeId:         nodeId,
						recordSliceLen: recordSliceLen,
						editDist:       an.editDist,
					})
				} else {
					// Replace the next character to match the prefix, increasing the edit distance.
					replaceEditDist := an.editDist + editDistForReplacePrefixChar
					if replaceEditDist <= editDistThreshold {
						ans = append(ans, activeNode{
							nodeId:         nodeId,
							recordSliceLen: recordSliceLen,
							editDist:       replaceEditDist,
						})
					}
				}
			})
		}
	}

	// Use the smallest edit distance for each unique node.
	ans = deduplicateActiveNodes(ans)

	// Cache the active nodes for reuse in later queries.
	t.activeNodeCache[prefix] = ans

	return ans
}

func (t *trie) expandActiveNode(an activeNode, f func(nodeId int, recordSliceLen int)) {
	node := t.nodes[an.nodeId]
	if an.recordSliceLen < len(node.recordSlice) {
		f(an.nodeId, an.recordSliceLen+1)
	} else {
		for _, childNodeId := range node.childNodeIds {
			f(childNodeId, 1)
		}
	}
}

// trieNode represents a node in the trie.
type trieNode struct {
	// Slice of the record string for the edge into this node.
	// The root has an empty slice; all other nodes have at least one byte.
	recordSlice string

	childNodeIds uniqueSortedIntSlice
	recordIds    uniqueSortedIntSlice

	// Store the range of record IDs in this node and all its descendants.
	// This allows the walk function to filter children that
	// don't have a record of interest.
	minRecordId, maxRecordId int
}

func (tn *trieNode) updateMinAndMaxRecordId(recordId int) {
	if recordId < tn.minRecordId {
		tn.minRecordId = recordId
	}
	if recordId > tn.maxRecordId {
		tn.maxRecordId = recordId
	}
}

type uniqueSortedIntSlice []int

func (s *uniqueSortedIntSlice) add(x int) {
	// Binary search.
	i := sort.SearchInts(*s, x)
	if i < len(*s) && (*s)[i] == x {
		// Already in the set.
		return
	}

	// Insertion sort to preserve order (ascending by value).
	*s = append(*s, x)
	for i := len(*s) - 1; i > 0; i-- {
		if (*s)[i] < (*s)[i-1] {
			(*s)[i], (*s)[i-1] = (*s)[i-1], (*s)[i]
		} else if (*s)[i] == (*s)[i-1] {
			panic("duplicate")
		} else {
			break
		}
	}
}

// activeNode represents a prefix in the trie within editDistThreshold of a query prefix.
type activeNode struct {
	nodeId         int
	recordSliceLen int
	editDist       float64
}

func deduplicateActiveNodes(activeNodes []activeNode) []activeNode {
	sort.Slice(activeNodes, func(i, j int) bool {
		if activeNodes[i].nodeId != activeNodes[j].nodeId {
			return activeNodes[i].nodeId < activeNodes[j].nodeId
		} else {
			return activeNodes[i].editDist < activeNodes[j].editDist
		}
	})

	var i int
	for j := 1; j < len(activeNodes); j++ {
		if !(activeNodes[j].nodeId == activeNodes[i].nodeId && activeNodes[j].recordSliceLen == activeNodes[i].recordSliceLen) {
			i++
			activeNodes[i] = activeNodes[j]
		}
	}
	return activeNodes[0:i]
}

func sharedPrefixLen(x, y string) int {
	var n int
	for n < len(x) && n < len(y) && x[n] == y[n] {
		n++
	}
	return n
}
