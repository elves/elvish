// Package vector implements persistent vector.
//
// This is a Go clone of Clojure's PersistentVector type
// (https://github.com/clojure/clojure/blob/master/src/jvm/clojure/lang/PersistentVector.java).
// For an introduction to the internals, see
// https://hypirion.com/musings/understanding-persistent-vector-pt-1.
package vector

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const (
	chunkBits  = 5
	nodeSize   = 1 << chunkBits
	tailMaxLen = nodeSize
	chunkMask  = nodeSize - 1
)

// Vector is a persistent sequential container for arbitrary values. It supports
// O(1) lookup by index, modification by index, and insertion and removal
// operations at the end. Being a persistent variant of the data structure, it
// is immutable, and provides O(1) operations to create modified versions of the
// vector that shares the underlying data structure, making it suitable for
// concurrent access. The empty value is a valid empty vector.
type Vector interface {
	json.Marshaler
	// Len returns the length of the vector.
	Len() int
	// Index returns the i-th element of the vector, if it exists. The second
	// return value indicates whether the element exists.
	Index(i int) (any, bool)
	// Assoc returns an almost identical Vector, with the i-th element
	// replaced. If the index is smaller than 0 or greater than the length of
	// the vector, it returns nil. If the index is equal to the size of the
	// vector, it is equivalent to Conj.
	Assoc(i int, val any) Vector
	// Conj returns an almost identical Vector, with an additional element
	// appended to the end.
	Conj(val any) Vector
	// Pop returns an almost identical Vector, with the last element removed. It
	// returns nil if the vector is already empty.
	Pop() Vector
	// SubVector returns a subvector containing the elements from i up to but
	// not including j.
	SubVector(i, j int) Vector
	// Iterator returns an iterator over the vector.
	Iterator() Iterator
}

// Iterator is an iterator over vector elements. It can be used like this:
//
//	for it := v.Iterator(); it.HasElem(); it.Next() {
//	    elem := it.Elem()
//	    // do something with elem...
//	}
type Iterator interface {
	// Elem returns the element at the current position.
	Elem() any
	// HasElem returns whether the iterator is pointing to an element.
	HasElem() bool
	// Next moves the iterator to the next position.
	Next()
}

type vector struct {
	count int
	// height of the tree structure, defined to be 0 when root is a leaf.
	height uint
	root   node
	tail   []any
}

// Empty is an empty Vector.
var Empty Vector = &vector{}

// node is a node in the vector tree. It is always of the size nodeSize.
type node *[nodeSize]any

func newNode() node {
	return node(&[nodeSize]any{})
}

func clone(n node) node {
	a := *n
	return node(&a)
}

func nodeFromSlice(s []any) node {
	var n [nodeSize]any
	copy(n[:], s)
	return &n
}

// Count returns the number of elements in a Vector.
func (v *vector) Len() int {
	return v.count
}

// treeSize returns the number of elements stored in the tree (as opposed to the
// tail).
func (v *vector) treeSize() int {
	if v.count < tailMaxLen {
		return 0
	}
	return ((v.count - 1) >> chunkBits) << chunkBits
}

func (v *vector) Index(i int) (any, bool) {
	if i < 0 || i >= v.count {
		return nil, false
	}

	// The following is very similar to sliceFor, but is implemented separately
	// to avoid unnecessary copying.
	if i >= v.treeSize() {
		return v.tail[i&chunkMask], true
	}
	n := v.root
	for shift := v.height * chunkBits; shift > 0; shift -= chunkBits {
		n = n[(i>>shift)&chunkMask].(node)
	}
	return n[i&chunkMask], true
}

// sliceFor returns the slice where the i-th element is stored. The index must
// be in bound.
func (v *vector) sliceFor(i int) []any {
	if i >= v.treeSize() {
		return v.tail
	}
	n := v.root
	for shift := v.height * chunkBits; shift > 0; shift -= chunkBits {
		n = n[(i>>shift)&chunkMask].(node)
	}
	return n[:]
}

func (v *vector) Assoc(i int, val any) Vector {
	if i < 0 || i > v.count {
		return nil
	} else if i == v.count {
		return v.Conj(val)
	}
	if i >= v.treeSize() {
		newTail := append([]any(nil), v.tail...)
		newTail[i&chunkMask] = val
		return &vector{v.count, v.height, v.root, newTail}
	}
	return &vector{v.count, v.height, doAssoc(v.height, v.root, i, val), v.tail}
}

// doAssoc returns an almost identical tree, with the i-th element replaced by
// val.
func doAssoc(height uint, n node, i int, val any) node {
	m := clone(n)
	if height == 0 {
		m[i&chunkMask] = val
	} else {
		sub := (i >> (height * chunkBits)) & chunkMask
		m[sub] = doAssoc(height-1, m[sub].(node), i, val)
	}
	return m
}

func (v *vector) Conj(val any) Vector {
	// Room in tail?
	if v.count-v.treeSize() < tailMaxLen {
		newTail := make([]any, len(v.tail)+1)
		copy(newTail, v.tail)
		newTail[len(v.tail)] = val
		return &vector{v.count + 1, v.height, v.root, newTail}
	}
	// Full tail; push into tree.
	tailNode := nodeFromSlice(v.tail)
	newHeight := v.height
	var newRoot node
	// Overflow root?
	if (v.count >> chunkBits) > (1 << (v.height * chunkBits)) {
		newRoot = newNode()
		newRoot[0] = v.root
		newRoot[1] = newPath(v.height, tailNode)
		newHeight++
	} else {
		newRoot = v.pushTail(v.height, v.root, tailNode)
	}
	return &vector{v.count + 1, newHeight, newRoot, []any{val}}
}

// pushTail returns a tree with tail appended.
func (v *vector) pushTail(height uint, n node, tail node) node {
	if height == 0 {
		return tail
	}
	idx := ((v.count - 1) >> (height * chunkBits)) & chunkMask
	m := clone(n)
	child := n[idx]
	if child == nil {
		m[idx] = newPath(height-1, tail)
	} else {
		m[idx] = v.pushTail(height-1, child.(node), tail)
	}
	return m
}

// newPath creates a left-branching tree of specified height and leaf.
func newPath(height uint, leaf node) node {
	if height == 0 {
		return leaf
	}
	ret := newNode()
	ret[0] = newPath(height-1, leaf)
	return ret
}

func (v *vector) Pop() Vector {
	switch v.count {
	case 0:
		return nil
	case 1:
		return Empty
	}
	if v.count-v.treeSize() > 1 {
		newTail := make([]any, len(v.tail)-1)
		copy(newTail, v.tail)
		return &vector{v.count - 1, v.height, v.root, newTail}
	}
	newTail := v.sliceFor(v.count - 2)
	newRoot := v.popTail(v.height, v.root)
	newHeight := v.height
	if v.height > 0 && newRoot[1] == nil {
		newRoot = newRoot[0].(node)
		newHeight--
	}
	return &vector{v.count - 1, newHeight, newRoot, newTail}
}

// popTail returns a new tree with the last leaf removed.
func (v *vector) popTail(level uint, n node) node {
	idx := ((v.count - 2) >> (level * chunkBits)) & chunkMask
	if level > 1 {
		newChild := v.popTail(level-1, n[idx].(node))
		if newChild == nil && idx == 0 {
			return nil
		}
		m := clone(n)
		if newChild == nil {
			// This is needed since `m[idx] = newChild` would store an
			// interface{} with a non-nil type part, which is non-nil
			m[idx] = nil
		} else {
			m[idx] = newChild
		}
		return m
	} else if idx == 0 {
		return nil
	} else {
		m := clone(n)
		m[idx] = nil
		return m
	}
}

func (v *vector) SubVector(begin, end int) Vector {
	if begin < 0 || begin > end || end > v.count {
		return nil
	}
	return &subVector{v, begin, end}
}

func (v *vector) Iterator() Iterator {
	return newIterator(v)
}

func (v *vector) MarshalJSON() ([]byte, error) {
	return marshalJSON(v.Iterator())
}

type subVector struct {
	v     *vector
	begin int
	end   int
}

func (s *subVector) Len() int {
	return s.end - s.begin
}

func (s *subVector) Index(i int) (any, bool) {
	if i < 0 || s.begin+i >= s.end {
		return nil, false
	}
	return s.v.Index(s.begin + i)
}

func (s *subVector) Assoc(i int, val any) Vector {
	if i < 0 || s.begin+i > s.end {
		return nil
	} else if s.begin+i == s.end {
		return s.Conj(val)
	}
	return s.v.Assoc(s.begin+i, val).SubVector(s.begin, s.end)
}

func (s *subVector) Conj(val any) Vector {
	return s.v.Assoc(s.end, val).SubVector(s.begin, s.end+1)
}

func (s *subVector) Pop() Vector {
	switch s.Len() {
	case 0:
		return nil
	case 1:
		return Empty
	default:
		return s.v.SubVector(s.begin, s.end-1)
	}
}

func (s *subVector) SubVector(i, j int) Vector {
	return s.v.SubVector(s.begin+i, s.begin+j)
}

func (s *subVector) Iterator() Iterator {
	return newIteratorWithRange(s.v, s.begin, s.end)
}

func (s *subVector) MarshalJSON() ([]byte, error) {
	return marshalJSON(s.Iterator())
}

type iterator struct {
	v        *vector
	treeSize int
	index    int
	end      int
	path     []pathEntry
}

type pathEntry struct {
	node  node
	index int
}

func (e pathEntry) current() any {
	return e.node[e.index]
}

func newIterator(v *vector) *iterator {
	return newIteratorWithRange(v, 0, v.Len())
}

func newIteratorWithRange(v *vector, begin, end int) *iterator {
	it := &iterator{v, v.treeSize(), begin, end, nil}
	if it.index >= it.treeSize {
		return it
	}
	// Find the node for begin, remembering all nodes along the path.
	n := v.root
	for shift := v.height * chunkBits; shift > 0; shift -= chunkBits {
		idx := (begin >> shift) & chunkMask
		it.path = append(it.path, pathEntry{n, idx})
		n = n[idx].(node)
	}
	it.path = append(it.path, pathEntry{n, begin & chunkMask})
	return it
}

func (it *iterator) Elem() any {
	if it.index >= it.treeSize {
		return it.v.tail[it.index-it.treeSize]
	}
	return it.path[len(it.path)-1].current()
}

func (it *iterator) HasElem() bool {
	return it.index < it.end
}

func (it *iterator) Next() {
	if it.index+1 >= it.treeSize {
		// Next element is in tail. Just increment the index.
		it.index++
		return
	}
	// Find the deepest level that can be advanced.
	var i int
	for i = len(it.path) - 1; i >= 0; i-- {
		e := it.path[i]
		if e.index+1 < len(e.node) {
			break
		}
	}
	if i == -1 {
		panic("cannot advance; vector iterator bug")
	}
	// Advance on this node, and re-populate all deeper levels.
	it.path[i].index++
	for i++; i < len(it.path); i++ {
		it.path[i] = pathEntry{it.path[i-1].current().(node), 0}
	}
	it.index++
}

type marshalError struct {
	index int
	cause error
}

func (err *marshalError) Error() string {
	return fmt.Sprintf("element %d: %s", err.index, err.cause)
}

func marshalJSON(it Iterator) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('[')
	index := 0
	for ; it.HasElem(); it.Next() {
		if index > 0 {
			buf.WriteByte(',')
		}
		elemBytes, err := json.Marshal(it.Elem())
		if err != nil {
			return nil, &marshalError{index, err}
		}
		buf.Write(elemBytes)
		index++
	}
	buf.WriteByte(']')
	return buf.Bytes(), nil
}
