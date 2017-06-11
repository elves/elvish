// Package vector implements persistent vector.
package vector

const (
	chunkBits  = 5
	nodeSize   = 1 << chunkBits
	tailMaxLen = nodeSize
	chunkMask  = nodeSize - 1
)

// Vector is a persistent sequential container for arbitrary values. It supports
// O(1) lookup by index, modification by index, and insertion and removal
// operations at the end. Being a persistent variant of the data structure, it
// is immutable, and provides operations to create modified versions of the
// vector that shares the underlying data structure, making it suitable for
// concurrent access. The empty value is a valid empty vector.
type Vector interface {
	// Len returns the length of the vector.
	Len() int
	// Nth returns the i-th element of the vector.
	Nth(i int) interface{}
	// AssocN returns an almost identical Vector, with the i-th element
	// replaced. If the index is smaller than 0 or greater than the length of
	// the vector, it returns nil. If the index is equal to the size of the
	// vector, it is equivalent to Cons.
	AssocN(i int, val interface{}) Vector
	// Cons returns an almost identical Vector, with an additional element
	// appended to the end.
	Cons(val interface{}) Vector
	// Pop returns an almost identical Vector, with the last element removed.
	Pop() Vector
	// SubVector returns a subvector containing the elements from i up to but
	// not including j.
	SubVector(i, j int) Vector
}

type vector struct {
	count int
	// height of the tree structure, defined to be 0 when root is a leaf.
	height uint
	root   node
	tail   []interface{}
}

// Empty is an empty Vector.
var Empty Vector = &vector{}

// node is a node in the vector tree. It is always of the size nodeSize.
type node []interface{}

func newNode() node {
	return node(make([]interface{}, nodeSize))
}

func (n node) clone() node {
	m := newNode()
	copy(m, n)
	return m
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

// sliceFor returns the slice where the i-th element is stored. It returns nil
// if the index is out of bound.
func (v *vector) sliceFor(i int) []interface{} {
	if i < 0 || i >= v.count {
		return nil
	}
	if i >= v.treeSize() {
		return v.tail
	}
	n := v.root
	for shift := v.height * chunkBits; shift > 0; shift -= chunkBits {
		n = n[(i>>shift)&chunkMask].(node)
	}
	return n
}

func (v *vector) Nth(i int) interface{} {
	return v.sliceFor(i)[i&chunkMask]
}

func (v *vector) AssocN(i int, val interface{}) Vector {
	if i < 0 || i > v.count {
		return nil
	} else if i == v.count {
		return v.Cons(val)
	}
	if i >= v.treeSize() {
		newTail := append([]interface{}(nil), v.tail...)
		copy(newTail, v.tail)
		newTail[i&chunkMask] = val
		return &vector{v.count, v.height, v.root, newTail}
	}
	return &vector{v.count, v.height, doAssoc(v.height, v.root, i, val), v.tail}
}

// doAssoc returns an almost identical tree, with the i-th element replaced by
// val.
func doAssoc(height uint, n node, i int, val interface{}) node {
	m := n.clone()
	if height == 0 {
		m[i&chunkMask] = val
	} else {
		sub := (i >> (height * chunkBits)) & chunkMask
		m[sub] = doAssoc(height-1, m[sub].(node), i, val)
	}
	return m
}

func (v *vector) Cons(val interface{}) Vector {
	// Room in tail?
	if v.count-v.treeSize() < tailMaxLen {
		newTail := make([]interface{}, len(v.tail)+1)
		copy(newTail, v.tail)
		newTail[len(v.tail)] = val
		return &vector{v.count + 1, v.height, v.root, newTail}
	}
	// Full tail; push into tree.
	tailNode := node(v.tail)
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
	return &vector{v.count + 1, newHeight, newRoot, []interface{}{val}}
}

// pushTail returns a tree with tail appended.
func (v *vector) pushTail(height uint, n node, tail node) node {
	if height == 0 {
		return tail
	}
	idx := ((v.count - 1) >> (height * chunkBits)) & chunkMask
	m := n.clone()
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
		newTail := make([]interface{}, len(v.tail)-1)
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
		m := n.clone()
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
		m := n.clone()
		m[idx] = nil
		return m
	}
}

func (v *vector) SubVector(begin, end int) Vector {
	return &subVector{v, begin, end}
}

type subVector struct {
	v     *vector
	begin int
	end   int
}

func (s *subVector) Len() int {
	return s.end - s.begin
}

func (s *subVector) Nth(i int) interface{} {
	return s.v.Nth(s.begin + i)
}

func (s *subVector) AssocN(i int, val interface{}) Vector {
	if s.begin+i > s.end {
		return nil
	} else if s.begin+i == s.end {
		return s.Cons(val)
	}
	return s.v.AssocN(s.begin+i, val).SubVector(s.begin, s.end)
}

func (s *subVector) Cons(val interface{}) Vector {
	return s.v.AssocN(s.end, val).SubVector(s.begin, s.end+1)
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
