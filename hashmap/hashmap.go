// Package hashmap implements persistent hashmap.
package hashmap

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
	// Len returns the length of the hashmap.
	Len() int
	// Get returns whether there is a value associated with the given key, and
	// that value or nil.
	Get(k Key) (bool, interface{})
	// Assoc returns an almost identical hashmap, with the given key associated
	// with the given value.
	Assoc(k Key, v interface{}) HashMap
	// Without returns an almost identical hashmap, with the given key
	// associated with no value.
	Without(k Key) HashMap
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

func (m *hashMap) Get(k Key) (bool, interface{}) {
	return m.root.find(0, k.Hash(), k)
}

func (m *hashMap) Assoc(k Key, v interface{}) HashMap {
	added, newRoot := m.root.assoc(0, k.Hash(), k, v)
	newCount := m.count
	if added {
		newCount++
	}
	return &hashMap{newCount, newRoot}
}

func (m *hashMap) Without(k Key) HashMap {
	deleted, newRoot := m.root.without(0, k.Hash(), k)
	newCount := m.count
	if deleted {
		newCount--
	}
	return &hashMap{newCount, newRoot}
}

// node is an interface for all nodes in the hash map tree.
type node interface {
	// assoc adds a new pair of key and value. It returns whether the key did
	// not exist before (i.e. a new pair has been added, instead of replaced),
	// and the new node.
	assoc(shift, hash uint32, k Key, v interface{}) (bool, node)
	// without removes a key. It returns whether the key did not exist before
	// (i.e. a key was indeed removed) and the new node.
	without(shift, hash uint32, k Key) (bool, node)
	// find finds the value for a key. It returns whether such a pair exists,
	// and the found value.
	find(shift, hash uint32, k Key) (bool, interface{})
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

func (n *arrayNode) assoc(shift, hash uint32, k Key, v interface{}) (bool, node) {
	idx := chunk(shift, hash)
	child := n.children[idx]
	if child == nil {
		_, newChild := emptyBitmapNode.assoc(shift+chunkBits, hash, k, v)
		return true, n.withNewChild(idx, newChild, 1)
	}
	added, newChild := child.assoc(shift+chunkBits, hash, k, v)
	return added, n.withNewChild(idx, newChild, 0)
}

func (n *arrayNode) without(shift, hash uint32, k Key) (bool, node) {
	idx := chunk(shift, hash)
	child := n.children[idx]
	if child == nil {
		return false, n
	}
	_, newChild := child.without(shift+chunkBits, hash, k)
	if newChild == child {
		return false, n
	}
	if newChild == nil {
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

func (n *arrayNode) find(shift, hash uint32, k Key) (bool, interface{}) {
	idx := chunk(shift, hash)
	child := n.children[idx]
	if child == nil {
		return false, nil
	}
	return child.find(shift+chunkBits, hash, k)
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
	key   Key
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

func createNode(shift uint32, k1 Key, v1 interface{}, h2 uint32, k2 Key, v2 interface{}) node {
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

func replaceEntry(entries []mapEntry, i uint32, k Key, v interface{}) []mapEntry {
	newEntries := append([]mapEntry(nil), entries...)
	newEntries[i] = mapEntry{k, v}
	return newEntries
}

func (n *bitmapNode) assoc(shift, hash uint32, k Key, v interface{}) (bool, node) {
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

func (n *bitmapNode) without(shift, hash uint32, k Key) (bool, node) {
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

func (n *bitmapNode) find(shift, hash uint32, k Key) (bool, interface{}) {
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

type collisionNode struct {
	hash    uint32
	entries []mapEntry
}

func (n *collisionNode) assoc(shift, hash uint32, k Key, v interface{}) (bool, node) {
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

func (n *collisionNode) without(shift, hash uint32, k Key) (bool, node) {
	idx := n.findIndex(k)
	if idx == -1 {
		return false, n
	}
	if len(n.entries) == 1 {
		return true, nil
	}
	return true, &collisionNode{n.hash, withoutEntry(n.entries, uint32(idx))}
}

func (n *collisionNode) find(shift, hash uint32, k Key) (bool, interface{}) {
	idx := n.findIndex(k)
	if idx == -1 {
		return false, nil
	}
	return true, n.entries[idx].value
}

func (n *collisionNode) findIndex(k Key) int {
	for i, entry := range n.entries {
		if k.Equal(entry.key) {
			return i
		}
	}
	return -1
}
