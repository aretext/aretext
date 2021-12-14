package fuzzy

import (
	"container/heap"
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
	nodes           []trieNode               // Trie nodes are identified by index in this slice.
	activeNodeCache map[string]activeNodeSet // Query prefix -> activeNodeSet
}

// newTrie returns an empty trie.
func newTrie() *trie {
	var rootNode trieNode
	return &trie{
		nodes:           []trieNode{rootNode},
		activeNodeCache: make(map[string]activeNodeSet, 0),
	}
}

// insert adds a new string to the trie and associates it with a record ID.
func (t *trie) insert(s string, recordId int) {
	var currentNodeId int // Start at the root.
	for i := 0; i < len(s); i++ {
		char := s[i]
		if childIdx, ok := t.nodes[currentNodeId].childIdxForChar(char); ok {
			// Child already exists for this char, so follow it.
			child := &(t.nodes[currentNodeId].children[childIdx])
			if recordId < child.minRecordId {
				child.minRecordId = recordId
			}
			if recordId > child.maxRecordId {
				child.maxRecordId = recordId
			}
			currentNodeId = child.nodeId
		} else {
			// Child does not yet exist for this char, so add it.
			childNodeId := len(t.nodes)
			t.nodes = append(t.nodes, trieNode{})
			t.nodes[currentNodeId].insertChild(trieNodeChild{
				nodeId:      childNodeId,
				char:        char,
				minRecordId: recordId,
				maxRecordId: recordId,
			})
			currentNodeId = childNodeId
		}
	}
	t.nodes[currentNodeId].addRecordId(recordId)
}

// topRecordIdsForPrefix finds the closest strings in the trie to a given prefix,
// as determined by edit distance.
// If prevRecordIds is non-nil, it includes only records that are in prevRecordIds.
// The limit parameter controls the maximum number of records returned.
func (t *trie) topRecordIdsForPrefix(prefix string, prevRecordIds *recordIdSet, limit int) *recordIdSet {
	visitedNodeIds := make(map[int]struct{}, 0)
	ans := t.activeNodeSetForPrefix(prefix)
	pq := newPriorityQueueForActiveNodeSet(ans) // Min heap by edit distance, then nodeId.
	topRecordIds := newRecordIdSet(limit)
	for pq.Len() > 0 {
		pqi := heap.Pop(&pq).(priorityQueueItem)
		if _, ok := visitedNodeIds[pqi.nodeId]; ok {
			// If we have already visited this node, then any subsequent visit must have a greater edit distance.
			// (All subsequent visits correspond to an item in the priority queue or a descendant of an item
			// in the priority queue. The heap property guarantees that the items in the priority queue have
			// edit distance greater than or equal to the current item,
			// and since edit distance increases for each descendant, all the descendants also have
			// edit distances greater than or equal to the current item.)
			continue
		}

		node := t.nodes[pqi.nodeId]
		for _, recordId := range node.recordIds {
			if prevRecordIds != nil && !prevRecordIds.contains(recordId) {
				// Filter by prevRecordIds.
				continue
			}
			topRecordIds.add(recordId)
			if topRecordIds.length() >= limit {
				// Terminate if we've found enough records.
				return topRecordIds
			}
		}

		visitedNodeIds[pqi.nodeId] = struct{}{}
		childEditDist := pqi.editDist + editDistPerSuffixChar
		for _, child := range node.children {
			if prevRecordIds != nil && (prevRecordIds.max < child.minRecordId || prevRecordIds.min > child.maxRecordId) {
				// If the child doesn't have any records that match the filter, skip it.
				// This is a performance optimization, not required for correctness.
				continue
			}
			heap.Push(&pq, priorityQueueItem{
				nodeId:   child.nodeId,
				editDist: childEditDist,
			})
		}
	}
	return topRecordIds
}

// activeNodeSetForPrefix calculates the active nodes in the trie
// within the edit distance threshold of a query prefix.
// When possible, it reuses cached results from prior queries.
func (t *trie) activeNodeSetForPrefix(prefix string) activeNodeSet {
	if ans, ok := t.activeNodeCache[prefix]; ok {
		// Re-use active nodes calculated from an earlier query.
		return ans
	}

	ans := make(activeNodeSet, 0)

	// Helper function to add all nodes in a subtree rooted at nodeId that are within the edit distance threshold.
	addDescendantsWithinEditDistThreshold := func(rootNodeId int, rootEditDist float64) {
		type stackItem struct {
			nodeId   int
			editDist float64
		}
		var current stackItem
		stack := []stackItem{{rootNodeId, rootEditDist}}
		for len(stack) > 0 {
			current, stack = stack[len(stack)-1], stack[0:len(stack)-1]
			if current.editDist <= editDistThreshold {
				ans.insertIfMinEditDist(current.nodeId, current.editDist)
				for _, child := range t.nodes[current.nodeId].children {
					stack = append(stack, stackItem{
						nodeId:   child.nodeId,
						editDist: current.editDist + editDistForDeletePrefixChar,
					})
				}
			}
		}
	}

	// Recursively build the active node set from an empty string,
	// adding one character from the prefix at a time.
	if len(prefix) == 0 {
		// Base case: empty prefix.
		addDescendantsWithinEditDistThreshold(0, 0)
	} else {
		// Recursive case: expand the active nodes by one character.
		prevActiveNodeSet := t.activeNodeSetForPrefix(prefix[0 : len(prefix)-1])
		for nodeId, editDist := range prevActiveNodeSet {
			// This case represents deleting the next character from the query,
			// which increases the edit distance.
			if deletionEditDist := editDist + editDistForDeleteQueryChar; deletionEditDist <= editDistThreshold {
				ans.insertIfMinEditDist(nodeId, deletionEditDist)
			}

			for _, child := range t.nodes[nodeId].children {
				if child.char == prefix[len(prefix)-1] {
					// The next character matches the prefix, so the child is an
					// active node with the same edit distance as the parent.
					// We also need to include all descendants of the child within
					// the edit distance threshold.
					addDescendantsWithinEditDistThreshold(child.nodeId, editDist)
				} else if replaceEditDist := editDist + editDistForReplacePrefixChar; replaceEditDist <= editDistThreshold {
					// Replace the next character to match the prefix,
					// which increases the edit distance.
					ans.insertIfMinEditDist(child.nodeId, replaceEditDist)
				}
			}
		}
	}

	t.activeNodeCache[prefix] = ans
	return ans
}

// trieNode represents a node in the trie.
type trieNode struct {
	children  []trieNodeChild // Sorted ascending by char.
	recordIds []int
}

func (tn *trieNode) insertChild(child trieNodeChild) {
	tn.children = append(tn.children, child)

	// Insertion-sort the new child to preserve the order (ascending by char).
	for i := len(tn.children) - 1; i > 0; i-- {
		if tn.children[i].char < tn.children[i-1].char {
			tn.children[i], tn.children[i-1] = tn.children[i-1], tn.children[i]
		} else if tn.children[i].char == tn.children[i-1].char {
			panic("Duplicate child")
		} else {
			break
		}
	}
}

func (tn trieNode) childIdxForChar(char byte) (int, bool) {
	// Binary search by char.
	i := sort.Search(len(tn.children), func(i int) bool {
		return tn.children[i].char >= char
	})
	if i < len(tn.children) && tn.children[i].char == char {
		return i, true
	} else {
		return 0, false
	}
}

func (tn *trieNode) addRecordId(recordId int) {
	// Binary search by record ID.
	i := sort.SearchInts(tn.recordIds, recordId)
	if i < len(tn.recordIds) && tn.recordIds[i] == recordId {
		// Record ID is already in the set
		return
	}

	// Insertion sort the record ID to preserve order (ascending by ID).
	tn.recordIds = append(tn.recordIds, recordId)
	for i := len(tn.recordIds) - 1; i > 0; i-- {
		if tn.recordIds[i] < tn.recordIds[i-1] {
			tn.recordIds[i], tn.recordIds[i-1] = tn.recordIds[i-1], tn.recordIds[i]
		} else if tn.recordIds[i] == tn.recordIds[i-1] {
			panic("Duplicate record")
		} else {
			break
		}
	}
}

// trieNodeChild represents an edge from one trie node to its child.
type trieNodeChild struct {
	nodeId int
	char   byte

	// Store the range of record IDs in this child and all its descendants.
	// This allows the walk function to filter children that
	// don't have a record of interest.
	minRecordId, maxRecordId int
}

// Set of active nodes and their edit distances.
// The key is an index into the the trie's nodes slice,
// and the value is the edit distance.
type activeNodeSet map[int]float64

func (ans activeNodeSet) insertIfMinEditDist(nodeId int, editDist float64) {
	currentEditDist, ok := ans[nodeId]
	if !ok || editDist < currentEditDist {
		ans[nodeId] = editDist
	}
}

// priorityQueue is a min-heap based on edit distance.
// It satisfies heap.Interface in the standard library.
type priorityQueue []priorityQueueItem

type priorityQueueItem struct {
	nodeId   int
	editDist float64
}

func newPriorityQueueForActiveNodeSet(ans activeNodeSet) priorityQueue {
	pq := make(priorityQueue, 0, len(ans))
	for nodeId, editDist := range ans {
		pq = append(pq, priorityQueueItem{
			nodeId:   nodeId,
			editDist: editDist,
		})
	}
	heap.Init(&pq)
	return pq
}

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	if pq[i].editDist != pq[j].editDist {
		return pq[i].editDist < pq[j].editDist
	} else {
		return pq[i].nodeId < pq[j].nodeId
	}
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(priorityQueueItem))
}

func (pq *priorityQueue) Pop() interface{} {
	n := len(*pq)
	x := (*pq)[n-1]
	*pq = (*pq)[0 : n-1]
	return x
}
