package persistent

type vectorNode []interface{}

func newVectorNode() vectorNode {
	return vectorNode(make([]interface{}, nodeCap))
}

func (n vectorNode) clone() vectorNode {
	m := newVectorNode()
	copy(m, n)
	return m
}

// Vector is an implementation of persistent vector. &Vector{} is a valid empty
// vector.
type Vector struct {
	count uint
	// height of the tree structure, defined to be 0 when root is a leaf.
	height uint
	root   vectorNode
	tail   []interface{}
}

var emptyVector = &Vector{}

// Count returns the number of elements in a Vector.
func (v *Vector) Count() uint {
	return v.count
}

// tailoff returns the number of elements not stored in tail.
func (v *Vector) tailoff() uint {
	if v.count < tailMaxLen {
		return 0
	}
	return ((v.count - 1) >> bitChunk) << bitChunk
}

// sliceFor returns the slice where the i-th element is stored.
func (v *Vector) sliceFor(i uint) []interface{} {
	if i < 0 || i >= v.count {
		return nil
	}
	if i >= v.tailoff() {
		return v.tail
	}
	n := v.root
	for shift := v.height * bitChunk; shift > 0; shift -= bitChunk {
		n = n[(i>>shift)&mask].(vectorNode)
	}
	return n
}

// Nth returns the i-th element.
func (v *Vector) Nth(i uint) interface{} {
	return v.sliceFor(i)[i&mask]
}

// AssocN returns a new Vector with the i-th element replaced by val.
func (v *Vector) AssocN(i uint, val interface{}) *Vector {
	if i < 0 || i > v.count {
		return nil
	} else if i == v.count {
		return v.Cons(val)
	}
	if i >= v.tailoff() {
		newTail := make([]interface{}, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&mask] = val
		return &Vector{v.count, v.height, v.root, newTail}
	}
	return &Vector{v.count, v.height, doAssoc(v.height, v.root, i, val), v.tail}
}

// doAssoc returns a new tree with the i-th element replaced by val.
func doAssoc(height uint, n vectorNode, i uint, val interface{}) vectorNode {
	m := n.clone()
	if height == 0 {
		m[i&mask] = val
	} else {
		sub := (i >> (height * bitChunk)) & mask
		m[sub] = doAssoc(height-1, m[sub].(vectorNode), i, val)
	}
	return m
}

// Cons returns a new Vector with val appended to the end.
func (v *Vector) Cons(val interface{}) *Vector {
	// Room in tail?
	if v.count-v.tailoff() < tailMaxLen {
		newTail := make([]interface{}, len(v.tail)+1)
		copy(newTail, v.tail)
		newTail[len(v.tail)] = val
		return &Vector{v.count + 1, v.height, v.root, newTail}
	}
	// Full tail; push into tree.
	tailNode := vectorNode(v.tail)
	newHeight := v.height
	var newRoot vectorNode
	// Overflow root?
	if (v.count >> bitChunk) > (1 << (v.height * bitChunk)) {
		newRoot = newVectorNode()
		newRoot[0] = v.root
		newRoot[1] = newPath(v.height, tailNode)
		newHeight++
	} else {
		newRoot = v.pushTail(v.height, v.root, tailNode)
	}
	return &Vector{v.count + 1, newHeight, newRoot, []interface{}{val}}
}

// pushTail returns a tree with tail appended.
func (v *Vector) pushTail(height uint, n vectorNode, tail vectorNode) vectorNode {
	if height == 0 {
		return tail
	}
	idx := ((v.count - 1) >> (height * bitChunk)) & mask
	m := n.clone()
	child := n[idx]
	if child == nil {
		m[idx] = newPath(height-1, tail)
	} else {
		m[idx] = v.pushTail(height-1, child.(vectorNode), tail)
	}
	return m
}

// newPath creates a left-branching tree of specified height and leaf.
func newPath(height uint, leaf vectorNode) vectorNode {
	if height == 0 {
		return leaf
	}
	ret := newVectorNode()
	ret[0] = newPath(height-1, leaf)
	return ret
}

// Pop returns a new Vector with the last element removed.
func (v *Vector) Pop() *Vector {
	switch v.count {
	case 0:
		return nil
	case 1:
		return emptyVector
	}
	if v.count-v.tailoff() > 1 {
		newTail := make([]interface{}, len(v.tail)-1)
		copy(newTail, v.tail)
		return &Vector{v.count - 1, v.height, v.root, newTail}
	}
	newTail := v.sliceFor(v.count - 2)
	newRoot := v.popTail(v.height, v.root)
	newHeight := v.height
	if v.height > 0 && newRoot[1] == nil {
		newRoot = newRoot[0].(vectorNode)
		newHeight--
	}
	return &Vector{v.count - 1, newHeight, newRoot, newTail}
}

// popTail returns a new tree with the last leaf removed.
func (v *Vector) popTail(level uint, n vectorNode) vectorNode {
	idx := ((v.count - 2) >> (level * bitChunk)) & mask
	if level > 1 {
		newChild := v.popTail(level-1, n[idx].(vectorNode))
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
