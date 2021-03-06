// Package pairing implements a Pairing heap Data structure
//
// Structure is not thread safe.
//
// Reference: https://en.wikipedia.org/wiki/Pairing_heap
package pairing

import (
	heap "github.com/theodesp/go-heaps"
)

// PairHeap is an implementation of a Pairing Heap.
// The zero value for PairHeap Root is an empty Heap.
type PairHeap struct {
	root       *node
}

// node contains the current item and the list if the sub-heaps
type node struct {
	// for use by client; untouched by this library
	item heap.Item
	// List of children nodes all containing values less than the Top of the heap
	children []*node
	// A reference to the parent Heap Node
	parent *node
}

func (n *node) detach() []*node {
	if n.parent == nil {
		return nil // avoid detaching root
	}
	for _, node := range n.children {
		node.parent = n.parent
	}
	var idx int
	for i, node := range n.parent.children {
		if node == n {
			idx = i
			break
		}
	}
	n.parent.children = append(n.parent.children[:idx], n.parent.children[idx+1:]...)
	n.parent = nil
	return n.children
}

func (n *node) findNode(item heap.Item) *node {
	if n.item.Compare(item) == 0 {
		return n
	} else {
		return n.findInChildren(n.children, item)
	}
}

func (n *node) findInChildren(children []*node, item heap.Item) *node {
	if len(children) == 0 {
		return nil
	}
	var node *node
loop:
	for _, child := range children {
		node = child.findNode(item)
		if node != nil {
			break loop
		}
	}
	return node
}

// Init initializes or clears the PairHeap
func (p *PairHeap) Init() *PairHeap {
	p.root = &node{}
	return p
}

// New returns an initialized PairHeap.
func New() *PairHeap { return new(PairHeap).Init() }

// IsEmpty returns true if PairHeap p is empty.
// The complexity is O(1).
func (p *PairHeap) IsEmpty() bool {
	return p.root.item == nil
}

// Resets the current PairHeap
func (p *PairHeap) Clear() {
	p.root = &node{}
}

// Find the smallest item in the priority queue.
// The complexity is O(1).
func (p *PairHeap) FindMin() heap.Item {
	if p.IsEmpty() {
		return nil
	}
	return p.root.item
}

// Inserts the value to the PairHeap and returns the item
// The complexity is O(1).
func (p *PairHeap) Insert(v heap.Item) heap.Item {
	n := node{item: v}
	merge(&p.root, &n)
	return n.item
}


// toDelete details what item to remove in a node call.
type toDelete int

const (
	removeItem toDelete = iota   // removes the given item
	removeMin                  // removes min item in the heap
)

// DeleteMin removes the top most value from the PairHeap and returns it
// The complexity is O(log n) amortized.
func (p *PairHeap) DeleteMin() heap.Item {
	return p.deleteItem(nil, removeMin)
}

// Deletes a node from the heap and returns the item
// The complexity is O(log n) amortized.
func (p *PairHeap) Delete(item heap.Item) heap.Item {
	return p.deleteItem(item, removeItem)
}

func (p *PairHeap) deleteItem(item heap.Item, typ toDelete) heap.Item {
	var result node

	switch typ {
	case removeMin:
		if len(p.root.children) == 0 {
			result = *p.root
			p.root.item = nil
		} else {
			result = *mergePairs(&p.root, p.root.children)
		}
	case removeItem:
		node := p.root.findNode(item)
		if node == nil {
			return nil
		} else {
			children := node.detach()
			result = *mergePairs(&p.root, append(p.root.children, children...))
		}
	default:
		panic("invalid type")
	}
	return result.item
}

// Adjusts the value to the node item and returns it
// The complexity is O(n) amortized.
func (p *PairHeap) Adjust(item heap.Item, new heap.Item) heap.Item {
	n := p.root.findNode(item)
	if n == nil {
		return nil
	}

	if n == p.root {
		p.DeleteMin()
		return p.Insert(new)
	} else {
		children := n.detach()
		n.item = new
		mergePairs(&p.root, append(p.root.children, append([]*node{n}, children...)...))
		return n.item
	}
}

// Do calls function cb on each element of the PairingHeap, in order of appearance.
// The behavior of Do is undefined if cb changes *p.
func (p *PairHeap) Do(cb func(item heap.Item)) {
	if p.IsEmpty() {
		return
	}
	// Call root first
	cb(p.root.item)
	// Then continue to children
	visitChildren(p.root.children, cb)
}

// Exhausting search of the element that matches item and returns it
// The complexity is O(n) amortized.
func (p *PairHeap) Find(item heap.Item) heap.Item {
	if p.IsEmpty() {
		return nil
	}
	node := p.root.findNode(item)
	if node == nil {
		return nil
	} else {
		return node.item
	}
}

func visitChildren(children []*node, cb func(item heap.Item)) {
	if len(children) == 0 {
		return
	}
	for _, heapNode := range children {
		cb(heapNode.item)
		visitChildren(heapNode.children, cb)
	}
}

func merge(first **node, second *node) *node {
	q := *first
	if q.item == nil { // Case when root is empty
		*first = second
		return *first
	}

	cmp := q.item.Compare(second.item)
	if cmp < 0 {
		// put 'second' as the first child of 'first' and update the parent
		q.children = append([]*node{second}, q.children...)
		second.parent = *first
		return *first
	} else {
		// put 'first' as the first child of 'second' and update the parent
		second.children = append([]*node{q}, second.children...)
		q.parent = second
		*first = second
		return second
	}
}

// Merges heaps together
func mergePairs(root **node, heaps []*node) *node {
	q := *root
	if len(heaps) == 1 {
		*root = heaps[0]
		heaps[0].parent = nil
		return q
	}
	var merged *node
	for { // iteratively merge heaps
		if len(heaps) == 0 {
			break
		}
		if merged == nil {
			merged = merge(&heaps[0], heaps[1])
			heaps = heaps[2:]
		} else {
			merged = merge(&merged, heaps[0])
			heaps = heaps[1:]
		}
	}
	*root = merged
	merged.parent = nil

	return q
}
