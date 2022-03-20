// Package hashmap implements persistent hashmap.
package hashmap

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

const (
	chunkBits = 5
	nodeCap   = 1 << chunkBits
	chunkMask = nodeCap - 1
)

// Equal is the type of a function that reports whether two keys are equal.
type Equal func(k1, k2 any) bool

// Hash is the type of a function that returns the hash code of a key.
type Hash func(k any) uint32

// New takes an equality function and a hash function, and returns an empty
// Map.
func New(e Equal, h Hash) Map {
	return &hashMap{0, emptyBitmapNode, nil, e, h}
}

type hashMap struct {
	count int
	root  node
	nilV  *any
	equal Equal
	hash  Hash
}

func (m *hashMap) Len() int {
	return m.count
}

func (m *hashMap) Index(k any) (any, bool) {
	if k == nil {
		if m.nilV == nil {
			return nil, false
		}
		return *m.nilV, true
	}
	return m.root.find(0, m.hash(k), k, m.equal)
}

func (m *hashMap) Assoc(k, v any) Map {
	if k == nil {
		newCount := m.count
		if m.nilV == nil {
			newCount++
		}
		return &hashMap{newCount, m.root, &v, m.equal, m.hash}
	}
	newRoot, added := m.root.assoc(0, m.hash(k), k, v, m.hash, m.equal)
	newCount := m.count
	if added {
		newCount++
	}
	return &hashMap{newCount, newRoot, m.nilV, m.equal, m.hash}
}

func (m *hashMap) Dissoc(k any) Map {
	if k == nil {
		newCount := m.count
		if m.nilV != nil {
			newCount--
		}
		return &hashMap{newCount, m.root, nil, m.equal, m.hash}
	}
	newRoot, deleted := m.root.without(0, m.hash(k), k, m.equal)
	newCount := m.count
	if deleted {
		newCount--
	}
	return &hashMap{newCount, newRoot, m.nilV, m.equal, m.hash}
}

func (m *hashMap) Iterator() Iterator {
	if m.nilV != nil {
		return &nilVIterator{true, *m.nilV, m.root.iterator()}
	}
	return m.root.iterator()
}

type nilVIterator struct {
	atNil bool
	nilV  any
	tail  Iterator
}

func (it *nilVIterator) Elem() (any, any) {
	if it.atNil {
		return nil, it.nilV
	}
	return it.tail.Elem()
}

func (it *nilVIterator) HasElem() bool {
	return it.atNil || it.tail.HasElem()
}

func (it *nilVIterator) Next() {
	if it.atNil {
		it.atNil = false
	} else {
		it.tail.Next()
	}
}

func (m *hashMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	for it := m.Iterator(); it.HasElem(); it.Next() {
		if first {
			first = false
		} else {
			buf.WriteByte(',')
		}
		k, v := it.Elem()
		kString, err := convertKey(k)
		if err != nil {
			return nil, err
		}
		kBytes, err := json.Marshal(kString)
		if err != nil {
			return nil, err
		}
		vBytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		buf.Write(kBytes)
		buf.WriteByte(':')
		buf.Write(vBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// convertKey converts a map key to a string. The implementation matches the
// behavior of how json.Marshal encodes keys of the builtin map type.
func convertKey(k any) (string, error) {
	kref := reflect.ValueOf(k)
	if kref.Kind() == reflect.String {
		return kref.String(), nil
	}
	if t, ok := k.(encoding.TextMarshaler); ok {
		b2, err := t.MarshalText()
		if err != nil {
			return "", err
		}
		return string(b2), nil
	}
	switch kref.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(kref.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(kref.Uint(), 10), nil
	}
	return "", fmt.Errorf("unsupported key type %T", k)
}

// node is an interface for all nodes in the hash map tree.
type node interface {
	// assoc adds a new pair of key and value. It returns the new node, and
	// whether the key did not exist before (i.e. a new pair has been added,
	// instead of replaced).
	assoc(shift, hash uint32, k, v any, h Hash, eq Equal) (node, bool)
	// without removes a key. It returns the new node and whether the key did
	// not exist before (i.e. a key was indeed removed).
	without(shift, hash uint32, k any, eq Equal) (node, bool)
	// find finds the value for a key. It returns the found value (if any) and
	// whether such a pair exists.
	find(shift, hash uint32, k any, eq Equal) (any, bool)
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

func (n *arrayNode) assoc(shift, hash uint32, k, v any, h Hash, eq Equal) (node, bool) {
	idx := chunk(shift, hash)
	child := n.children[idx]
	if child == nil {
		newChild, _ := emptyBitmapNode.assoc(shift+chunkBits, hash, k, v, h, eq)
		return n.withNewChild(idx, newChild, 1), true
	}
	newChild, added := child.assoc(shift+chunkBits, hash, k, v, h, eq)
	return n.withNewChild(idx, newChild, 0), added
}

func (n *arrayNode) without(shift, hash uint32, k any, eq Equal) (node, bool) {
	idx := chunk(shift, hash)
	child := n.children[idx]
	if child == nil {
		return n, false
	}
	newChild, _ := child.without(shift+chunkBits, hash, k, eq)
	if newChild == child {
		return n, false
	}
	if newChild == emptyBitmapNode {
		if n.nChildren <= nodeCap/4 {
			// less than 1/4 full; shrink
			return n.pack(int(idx)), true
		}
		return n.withNewChild(idx, nil, -1), true
	}
	return n.withNewChild(idx, newChild, 0), true
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

func (n *arrayNode) find(shift, hash uint32, k any, eq Equal) (any, bool) {
	idx := chunk(shift, hash)
	child := n.children[idx]
	if child == nil {
		return nil, false
	}
	return child.find(shift+chunkBits, hash, k, eq)
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

func (it *arrayNodeIterator) Elem() (any, any) {
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
	key   any
	value any
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
	m2  uint32 = 0x33333333
	m4  uint32 = 0x0f0f0f0f
	m8  uint32 = 0x00ff00ff
	m16 uint32 = 0x0000ffff
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

func createNode(shift uint32, k1 any, v1 any, h2 uint32, k2 any, v2 any, h Hash, eq Equal) node {
	h1 := h(k1)
	if h1 == h2 {
		return &collisionNode{h1, []mapEntry{{k1, v1}, {k2, v2}}}
	}
	n, _ := emptyBitmapNode.assoc(shift, h1, k1, v1, h, eq)
	n, _ = n.assoc(shift, h2, k2, v2, h, eq)
	return n
}

func (n *bitmapNode) unpack(shift, idx uint32, newChild node, h Hash, eq Equal) *arrayNode {
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
				newNode.children[i], _ = emptyBitmapNode.assoc(
					shift+chunkBits, h(entry.key), entry.key, entry.value, h, eq)
			}
		}
	}
	return &newNode
}

func (n *bitmapNode) withoutEntry(bit, idx uint32) *bitmapNode {
	if n.bitmap == bit {
		return emptyBitmapNode
	}
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

func replaceEntry(entries []mapEntry, i uint32, k, v any) []mapEntry {
	newEntries := append([]mapEntry(nil), entries...)
	newEntries[i] = mapEntry{k, v}
	return newEntries
}

func (n *bitmapNode) assoc(shift, hash uint32, k, v any, h Hash, eq Equal) (node, bool) {
	bit := bitpos(shift, hash)
	idx := index(n.bitmap, bit)
	if n.bitmap&bit == 0 {
		// Entry does not exist yet
		nEntries := len(n.entries)
		if nEntries >= nodeCap/2 {
			// Unpack into an arrayNode
			newNode, _ := emptyBitmapNode.assoc(shift+chunkBits, hash, k, v, h, eq)
			return n.unpack(shift, chunk(shift, hash), newNode, h, eq), true
		}
		// Add a new entry
		newEntries := make([]mapEntry, len(n.entries)+1)
		copy(newEntries[:idx], n.entries[:idx])
		newEntries[idx] = mapEntry{k, v}
		copy(newEntries[idx+1:], n.entries[idx:])
		return &bitmapNode{n.bitmap | bit, newEntries}, true
	}
	// Entry exists
	entry := n.entries[idx]
	if entry.key == nil {
		// Non-leaf child
		child := entry.value.(node)
		newChild, added := child.assoc(shift+chunkBits, hash, k, v, h, eq)
		return n.withReplacedEntry(idx, mapEntry{nil, newChild}), added
	}
	// Leaf
	if eq(k, entry.key) {
		// Identical key, replace
		return n.withReplacedEntry(idx, mapEntry{k, v}), false
	}
	// Create and insert new inner node
	newNode := createNode(shift+chunkBits, entry.key, entry.value, hash, k, v, h, eq)
	return n.withReplacedEntry(idx, mapEntry{nil, newNode}), true
}

func (n *bitmapNode) without(shift, hash uint32, k any, eq Equal) (node, bool) {
	bit := bitpos(shift, hash)
	if n.bitmap&bit == 0 {
		return n, false
	}
	idx := index(n.bitmap, bit)
	entry := n.entries[idx]
	if entry.key == nil {
		// Non-leaf child
		child := entry.value.(node)
		newChild, deleted := child.without(shift+chunkBits, hash, k, eq)
		if newChild == child {
			return n, false
		}
		if newChild == emptyBitmapNode {
			return n.withoutEntry(bit, idx), true
		}
		return n.withReplacedEntry(idx, mapEntry{nil, newChild}), deleted
	} else if eq(entry.key, k) {
		// Leaf, and this is the entry to delete.
		return n.withoutEntry(bit, idx), true
	}
	// Nothing to delete.
	return n, false
}

func (n *bitmapNode) find(shift, hash uint32, k any, eq Equal) (any, bool) {
	bit := bitpos(shift, hash)
	if n.bitmap&bit == 0 {
		return nil, false
	}
	idx := index(n.bitmap, bit)
	entry := n.entries[idx]
	if entry.key == nil {
		child := entry.value.(node)
		return child.find(shift+chunkBits, hash, k, eq)
	} else if eq(entry.key, k) {
		return entry.value, true
	}
	return nil, false
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

func (it *bitmapNodeIterator) Elem() (any, any) {
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

func (n *collisionNode) assoc(shift, hash uint32, k, v any, h Hash, eq Equal) (node, bool) {
	if hash == n.hash {
		idx := n.findIndex(k, eq)
		if idx != -1 {
			return &collisionNode{
				n.hash, replaceEntry(n.entries, uint32(idx), k, v)}, false
		}
		newEntries := make([]mapEntry, len(n.entries)+1)
		copy(newEntries[:len(n.entries)], n.entries[:])
		newEntries[len(n.entries)] = mapEntry{k, v}
		return &collisionNode{n.hash, newEntries}, true
	}
	// Wrap in a bitmapNode and add the entry
	wrap := bitmapNode{bitpos(shift, n.hash), []mapEntry{{nil, n}}}
	return wrap.assoc(shift, hash, k, v, h, eq)
}

func (n *collisionNode) without(shift, hash uint32, k any, eq Equal) (node, bool) {
	idx := n.findIndex(k, eq)
	if idx == -1 {
		return n, false
	}
	if len(n.entries) == 1 {
		return emptyBitmapNode, true
	}
	return &collisionNode{n.hash, withoutEntry(n.entries, uint32(idx))}, true
}

func (n *collisionNode) find(shift, hash uint32, k any, eq Equal) (any, bool) {
	idx := n.findIndex(k, eq)
	if idx == -1 {
		return nil, false
	}
	return n.entries[idx].value, true
}

func (n *collisionNode) findIndex(k any, eq Equal) int {
	for i, entry := range n.entries {
		if eq(k, entry.key) {
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

func (it *collisionNodeIterator) Elem() (any, any) {
	entry := it.n.entries[it.index]
	return entry.key, entry.value
}

func (it *collisionNodeIterator) HasElem() bool {
	return it.index < len(it.n.entries)
}

func (it *collisionNodeIterator) Next() {
	it.index++
}
