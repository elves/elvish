// Package hashmap implements persistent hashmap.
package hashmap

import "github.com/xiaq/persistent/types"

const (
	chunkBits = 5
	nodeCap   = 1 << chunkBits
	chunkMask = nodeCap - 1
)

// HashMap is a persistent associative data structure mapping keys to values. It
// is immutable, and supports near-O(1) operations to create modified version of
// the hashmap that shares the underlying data structure, making it suitable for
// concurrent access.
type HashMap interface {
	types.Equaler
	// Len returns the length of the hashmap.
	Len() int
	// Get returns whether there is a value associated with the given key, and
	// that value or nil.
	Get(k types.Key) (bool, interface{})
	// Assoc returns an almost identical hashmap, with the given key associated
	// with the given value.
	Assoc(k types.Key, v interface{}) HashMap
	// Without returns an almost identical hashmap, with the given key
	// associated with no value.
	Without(k types.Key) HashMap
	// Iterator returns an iterator over the map.
	Iterator() Iterator
}

// Iterator is an iterator over map elements. It can be used like this:
//
//     for it := m.Iterator(); it.HasElem(); it.Next() {
//         key, value := it.Elem()
//         // do something with elem...
//     }
type Iterator interface {
	// Elem returns the current key-value pair.
	Elem() (types.Key, interface{})
	// HasElem returns whether the iterator is pointing to an element.
	HasElem() bool
	// Next moves the iterator to the next position.
	Next()
}

// Empty is an empty hashmap.
var Empty HashMap = &hashMap{0, emptyBitmapNode}

type hashMap struct {
	count int
	root  node
}

func (m *hashMap) Len() int {
	return m.count
}

func (m *hashMap) Get(k types.Key) (bool, interface{}) {
	return m.root.find(0, k.Hash(), k)
}

func (m *hashMap) Assoc(k types.Key, v interface{}) HashMap {
	added, newRoot := m.root.assoc(0, k.Hash(), k, v)
	newCount := m.count
	if added {
		newCount++
	}
	return &hashMap{newCount, newRoot}
}

func (m *hashMap) Without(k types.Key) HashMap {
	deleted, newRoot := m.root.without(0, k.Hash(), k)
	newCount := m.count
	if deleted {
		newCount--
	}
	return &hashMap{newCount, newRoot}
}

func (m *hashMap) Iterator() Iterator {
	return m.root.iterator()
}

func (m *hashMap) Equal(other interface{}) bool {
	m2, ok := other.(HashMap)
	return ok && HashMapEqual(m, m2)
}

// HashMapEqual returns whether two HashMap values are structurally equal. The
// equality of the values are determined with m1[k].Equal(m2[k]) if m1[k]
// satisfies Equaler, or m1[k] == m2[k] otherwise. If there is any value in m1
// that do not satisfy Equaler and are uncomparable, this function can panic.
func HashMapEqual(m1, m2 HashMap) bool {
	if m1.Len() != m2.Len() {
		return false
	}
	for it := m1.Iterator(); it.HasElem(); it.Next() {
		k, v1 := it.Elem()
		ok2, v2 := m2.Get(k)
		if !ok2 {
			return false
		}
		if v1Eq, ok := v1.(types.Equaler); ok {
			if !v1Eq.Equal(v2) {
				return false
			}
		} else {
			if v1 != v2 {
				return false
			}
		}
	}
	return true
}

// node is an interface for all nodes in the hash map tree.
type node interface {
	// assoc adds a new pair of key and value. It returns whether the key did
	// not exist before (i.e. a new pair has been added, instead of replaced),
	// and the new node.
	assoc(shift, hash uint32, k types.Key, v interface{}) (bool, node)
	// without removes a key. It returns whether the key did not exist before
	// (i.e. a key was indeed removed) and the new node.
	without(shift, hash uint32, k types.Key) (bool, node)
	// find finds the value for a key. It returns whether such a pair exists,
	// and the found value.
	find(shift, hash uint32, k types.Key) (bool, interface{})
	// iterator returns an iterator.
	iterator() Iterator
}

// arrayNode stores all of its children in an array. The array is always at
// least 1/4 full, otherwise it will be packed into a bitmapNode.
type arrayNode struct {
	nChildren int
	children  [nodeCap]node
}

func (n *arrayNode) withNewChild(i uint32, newChild node, d int) *arrayNode {
	newChildren := n.children
	newChildren[i] = newChild
	return &arrayNode{n.nChildren + d, newChildren}
}

func (n *arrayNode) assoc(shift, hash uint32, k types.Key, v interface{}) (bool, node) {
	idx := chunk(shift, hash)
	child := n.children[idx]
	if child == nil {
		_, newChild := emptyBitmapNode.assoc(shift+chunkBits, hash, k, v)
		return true, n.withNewChild(idx, newChild, 1)
	}
	added, newChild := child.assoc(shift+chunkBits, hash, k, v)
	return added, n.withNewChild(idx, newChild, 0)
}

func (n *arrayNode) without(shift, hash uint32, k types.Key) (bool, node) {
	idx := chunk(shift, hash)
	child := n.children[idx]
	if child == nil {
		return false, n
	}
	_, newChild := child.without(shift+chunkBits, hash, k)
	if newChild == child {
		return false, n
	}
	if newChild == emptyBitmapNode {
		if n.nChildren <= nodeCap/4 {
			// less than 1/4 full; shrink
			return true, n.pack(int(idx))
		}
		return true, n.withNewChild(idx, nil, -1)
	}
	return true, n.withNewChild(idx, newChild, 0)
}

func (n *arrayNode) pack(skip int) *bitmapNode {
	newNode := bitmapNode{0, make([]mapEntry, n.nChildren-1)}
	j := 0
	for i, child := range n.children {
		// TODO(xiaq): Benchmark performance difference after unrolling this
		// into two loops without the if
		if i != skip && child != nil {
			newNode.bitmap |= 1 << uint(i)
			newNode.entries[j].value = child
			j++
		}
	}
	return &newNode
}

func (n *arrayNode) find(shift, hash uint32, k types.Key) (bool, interface{}) {
	idx := chunk(shift, hash)
	child := n.children[idx]
	if child == nil {
		return false, nil
	}
	return child.find(shift+chunkBits, hash, k)
}

func (n *arrayNode) iterator() Iterator {
	it := &arrayNodeIterator{n, 0, nil}
	it.fixCurrent()
	return it
}

type arrayNodeIterator struct {
	n       *arrayNode
	index   int
	current Iterator
}

func (it *arrayNodeIterator) fixCurrent() {
	for ; it.index < nodeCap && it.n.children[it.index] == nil; it.index++ {
	}
	if it.index < nodeCap {
		it.current = it.n.children[it.index].iterator()
	} else {
		it.current = nil
	}
}

func (it *arrayNodeIterator) Elem() (types.Key, interface{}) {
	return it.current.Elem()
}

func (it *arrayNodeIterator) HasElem() bool {
	return it.current != nil
}

func (it *arrayNodeIterator) Next() {
	it.current.Next()
	if !it.current.HasElem() {
		it.index++
		it.fixCurrent()
	}
}

var emptyBitmapNode = &bitmapNode{}

type bitmapNode struct {
	bitmap  uint32
	entries []mapEntry
}

// mapEntry is a map entry. When used in a collisionNode, it is also an entry
// with non-nil key. When used in a bitmapNode, it is also abused to represent
// children when the key is nil.
type mapEntry struct {
	key   types.Key
	value interface{}
}

func chunk(shift, hash uint32) uint32 {
	return (hash >> shift) & chunkMask
}

func bitpos(shift, hash uint32) uint32 {
	return 1 << chunk(shift, hash)
}

func index(bitmap, bit uint32) uint32 {
	return popCount(bitmap & (bit - 1))
}

const (
	m1  uint32 = 0x55555555
	m2         = 0x33333333
	m4         = 0x0f0f0f0f
	m8         = 0x00ff00ff
	m16        = 0x0000ffff
)

// TODO(xiaq): Use an optimized implementation.
func popCount(u uint32) uint32 {
	u = (u & m1) + ((u >> 1) & m1)
	u = (u & m2) + ((u >> 2) & m2)
	u = (u & m4) + ((u >> 4) & m4)
	u = (u & m8) + ((u >> 8) & m8)
	u = (u & m16) + ((u >> 16) & m16)
	return u
}

func createNode(shift uint32, k1 types.Key, v1 interface{}, h2 uint32, k2 types.Key, v2 interface{}) node {
	h1 := k1.Hash()
	if h1 == h2 {
		return &collisionNode{h1, []mapEntry{{k1, v1}, {k2, v2}}}
	}
	_, n := emptyBitmapNode.assoc(shift, h1, k1, v1)
	_, n = n.assoc(shift, h2, k2, v2)
	return n
}

func (n *bitmapNode) unpack(shift, idx uint32, newChild node) *arrayNode {
	var newNode arrayNode
	newNode.nChildren = len(n.entries) + 1
	newNode.children[idx] = newChild
	j := 0
	for i := uint(0); i < nodeCap; i++ {
		if (n.bitmap>>i)&1 != 0 {
			entry := n.entries[j]
			j++
			if entry.key == nil {
				newNode.children[i] = entry.value.(node)
			} else {
				_, newNode.children[i] = emptyBitmapNode.assoc(
					shift+chunkBits, entry.key.Hash(), entry.key, entry.value)
			}
		}
	}
	return &newNode
}

func (n *bitmapNode) withoutEntry(bit, idx uint32) *bitmapNode {
	return &bitmapNode{n.bitmap ^ bit, withoutEntry(n.entries, idx)}
}

func withoutEntry(entries []mapEntry, idx uint32) []mapEntry {
	newEntries := make([]mapEntry, len(entries)-1)
	copy(newEntries[:idx], entries[:idx])
	copy(newEntries[idx:], entries[idx+1:])
	return newEntries
}

func (n *bitmapNode) withReplacedEntry(i uint32, entry mapEntry) *bitmapNode {
	return &bitmapNode{n.bitmap, replaceEntry(n.entries, i, entry.key, entry.value)}
}

func replaceEntry(entries []mapEntry, i uint32, k types.Key, v interface{}) []mapEntry {
	newEntries := append([]mapEntry(nil), entries...)
	newEntries[i] = mapEntry{k, v}
	return newEntries
}

func (n *bitmapNode) assoc(shift, hash uint32, k types.Key, v interface{}) (bool, node) {
	bit := bitpos(shift, hash)
	idx := index(n.bitmap, bit)
	if n.bitmap&bit == 0 {
		// Entry does not exist yet
		nEntries := len(n.entries)
		if nEntries >= nodeCap/2 {
			// Unpack into an arrayNode
			_, newNode := emptyBitmapNode.assoc(shift+chunkBits, hash, k, v)
			return true, n.unpack(shift, chunk(shift, hash), newNode)
		}
		// Add a new entry
		newEntries := make([]mapEntry, len(n.entries)+1)
		copy(newEntries[:idx], n.entries[:idx])
		newEntries[idx] = mapEntry{k, v}
		copy(newEntries[idx+1:], n.entries[idx:])
		return true, &bitmapNode{n.bitmap | bit, newEntries}
	}
	// Entry exists
	entry := n.entries[idx]
	if entry.key == nil {
		// Non-leaf child
		child := entry.value.(node)
		added, newChild := child.assoc(shift+chunkBits, hash, k, v)
		return added, n.withReplacedEntry(idx, mapEntry{nil, newChild})
	}
	// Leaf
	if k.Equal(entry.key) {
		// Identical key, replace
		return false, n.withReplacedEntry(idx, mapEntry{k, v})
	}
	// Create and insert new inner node
	newNode := createNode(shift+chunkBits, entry.key, entry.value, hash, k, v)
	return true, n.withReplacedEntry(idx, mapEntry{nil, newNode})
}

func (n *bitmapNode) without(shift, hash uint32, k types.Key) (bool, node) {
	bit := bitpos(shift, hash)
	if n.bitmap&bit == 0 {
		return false, n
	}
	idx := index(n.bitmap, bit)
	entry := n.entries[idx]
	if entry.key == nil {
		// Non-leaf child
		child := entry.value.(node)
		deleted, newChild := child.without(shift+chunkBits, hash, k)
		if newChild == child {
			return false, n
		}
		if newChild == nil {
			// Sole element in subtree deleted
			if n.bitmap == bit {
				return true, emptyBitmapNode
			}
			return true, n.withoutEntry(bit, idx)
		}
		return deleted, n.withReplacedEntry(idx, mapEntry{nil, newChild})
	} else if entry.key.Equal(k) {
		// Leaf, and this is the entry to delete.
		return true, n.withoutEntry(bit, idx)
	}
	// Nothing to delete.
	return false, n
}

func (n *bitmapNode) find(shift, hash uint32, k types.Key) (bool, interface{}) {
	bit := bitpos(shift, hash)
	if n.bitmap&bit == 0 {
		return false, nil
	}
	idx := index(n.bitmap, bit)
	entry := n.entries[idx]
	if entry.key == nil {
		child := entry.value.(node)
		return child.find(shift+chunkBits, hash, k)
	} else if entry.key.Equal(k) {
		return true, entry.value
	}
	return false, nil
}

func (n *bitmapNode) iterator() Iterator {
	it := &bitmapNodeIterator{n, 0, nil}
	it.fixCurrent()
	return it
}

type bitmapNodeIterator struct {
	n       *bitmapNode
	index   int
	current Iterator
}

func (it *bitmapNodeIterator) fixCurrent() {
	if it.index < len(it.n.entries) {
		entry := it.n.entries[it.index]
		if entry.key == nil {
			it.current = entry.value.(node).iterator()
		} else {
			it.current = nil
		}
	} else {
		it.current = nil
	}
}

func (it *bitmapNodeIterator) Elem() (types.Key, interface{}) {
	if it.current != nil {
		return it.current.Elem()
	}
	entry := it.n.entries[it.index]
	return entry.key, entry.value
}

func (it *bitmapNodeIterator) HasElem() bool {
	return it.index < len(it.n.entries)
}

func (it *bitmapNodeIterator) Next() {
	if it.current != nil {
		it.current.Next()
	}
	if it.current == nil || !it.current.HasElem() {
		it.index++
		it.fixCurrent()
	}
}

type collisionNode struct {
	hash    uint32
	entries []mapEntry
}

func (n *collisionNode) assoc(shift, hash uint32, k types.Key, v interface{}) (bool, node) {
	if hash == n.hash {
		idx := n.findIndex(k)
		if idx != -1 {
			return false, &collisionNode{
				n.hash, replaceEntry(n.entries, uint32(idx), k, v)}
		}
		newEntries := make([]mapEntry, len(n.entries)+1)
		copy(newEntries[:len(n.entries)], n.entries[:])
		newEntries[len(n.entries)] = mapEntry{k, v}
		return true, &collisionNode{n.hash, newEntries}
	}
	// Wrap in a bitmapNode and add the entry
	wrap := bitmapNode{bitpos(shift, n.hash), []mapEntry{{nil, n}}}
	return wrap.assoc(shift, hash, k, v)
}

func (n *collisionNode) without(shift, hash uint32, k types.Key) (bool, node) {
	idx := n.findIndex(k)
	if idx == -1 {
		return false, n
	}
	if len(n.entries) == 1 {
		return true, nil
	}
	return true, &collisionNode{n.hash, withoutEntry(n.entries, uint32(idx))}
}

func (n *collisionNode) find(shift, hash uint32, k types.Key) (bool, interface{}) {
	idx := n.findIndex(k)
	if idx == -1 {
		return false, nil
	}
	return true, n.entries[idx].value
}

func (n *collisionNode) findIndex(k types.Key) int {
	for i, entry := range n.entries {
		if k.Equal(entry.key) {
			return i
		}
	}
	return -1
}

func (n *collisionNode) iterator() Iterator {
	return &collisionNodeIterator{n, 0}
}

type collisionNodeIterator struct {
	n     *collisionNode
	index int
}

func (it *collisionNodeIterator) Elem() (types.Key, interface{}) {
	entry := it.n.entries[it.index]
	return entry.key, entry.value
}

func (it *collisionNodeIterator) HasElem() bool {
	return it.index < len(it.n.entries)
}

func (it *collisionNodeIterator) Next() {
	it.index++
}
