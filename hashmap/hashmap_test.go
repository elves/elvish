package hashmap

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

// testKey is an implementation of the Key interface for testing.
type testKey uint64

// Hash returns the lower 32 bits. This is intended so that hash collisions can
// be easily constructed.
func (x testKey) Hash() uint32 {
	return uint32(x & 0xffffffff)
}

// Equal returns true if and only if the other value is also a testKey and they
// are equal.
func (x testKey) Equal(other interface{}) bool {
	y, ok := other.(testKey)
	return ok && x == y
}

type anotherTestKey struct{}

func (anotherTestKey) Hash() uint32           { return 0 }
func (anotherTestKey) Equal(interface{}) bool { return false }

const (
	NSequential = 0x1000
	NCollision  = 0x100
	NRandom     = 0x4000
)

type refEntry struct {
	k testKey
	v string
}

var refEntries = []refEntry{
	{0x100000001, "x1"},
	{0x200000001, "y1"},
}

func init() {
	add := func(k testKey, v string) {
		refEntries = append(refEntries, refEntry{k, v})
	}
	for i := 0; i < NSequential; i++ {
		add(testKey(i), hex(uint64(i)))
	}
	for i := 0; i < NCollision; i++ {
		add(testKey(uint64(i+1)<<32), "collision "+hex(uint64(i)))
	}
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < NRandom; i++ {
		// Avoid rand.Uint64 for compatibility with pre 1.8 Go
		k := uint64(rand.Int63())>>31 | uint64(rand.Int63())<<32
		add(testKey(k), "random "+hex(k))
	}
}

func hex(i uint64) string {
	return "0x" + strconv.FormatUint(i, 16)
}

func TestHashMap(t *testing.T) {
	m := Empty
	// Len of Empty should be 0.
	if m.Len() != 0 {
		t.Errorf("m.Len = %d, want %d", m.Len(), 0)
	}
	// Assoc and Len.
	size := 0
	for _, e := range refEntries {
		m = m.Assoc(e.k, e.v)
		size++
		if m.Len() != size {
			t.Errorf("m.Len = %d, want %d", m.Len(), size)
		}
	}
	// Build a reference map.
	ref := make(map[testKey]string, len(refEntries))
	for _, e := range refEntries {
		ref[e.k] = e.v
	}
	// Get.
	testMapContent(t, m, ref)
	in, got := m.Get(anotherTestKey{})
	if in {
		t.Errorf("m.Get <bad key> returns entry %v", got)
	}
	// Without.
	for _, e := range refEntries {
		delete(ref, e.k)
		m = m.Without(e.k)
		if m.Len() != len(ref) {
			t.Errorf("m.Len() = %d after removing, should be %v", m.Len(), len(ref))
		}
		in, _ := m.Get(e.k)
		if in {
			t.Errorf("m.Get(%v) still returns item after removal", e.k)
		}
		// Checking all elements is expensive. Only do this 1% of the time.
		if rand.Float64() < 0.01 {
			testMapContent(t, m, ref)
		}
	}
}

func testMapContent(t *testing.T, m HashMap, ref map[testKey]string) {
	for k, v := range ref {
		in, got := m.Get(k)
		if !in {
			t.Errorf("m.Get 0x%x returns no entry", k)
		}
		if got != v {
			t.Errorf("m.Get(0x%x) = %v, want %v", k, got, v)
		}
	}
}
