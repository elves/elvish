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

type anotherTestKey uint32

func (x anotherTestKey) Hash() uint32         { return uint32(x) }
func (anotherTestKey) Equal(interface{}) bool { return false }

const (
	NSequential = 0x1000
	NCollision  = 0x100
	NRandom     = 0x4000
	NReplace    = 0x200

	SmallRandomPass      = 0x100
	NSmallRandom         = 0x400
	SmallRandomHighBound = 0x50
	SmallRandomLowBound  = 0x200

	NArrayNode = 0x100

	NIneffectiveWithout = 0x200
)

type refEntry struct {
	k testKey
	v string
}

func hex(i uint64) string {
	return "0x" + strconv.FormatUint(i, 16)
}

func TestHashMap(t *testing.T) {
	var refEntries []refEntry
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
	for i := 0; i < NReplace; i++ {
		k := uint64(rand.Int31n(NSequential))
		add(testKey(k), "replace "+hex(k))
	}

	testHashMapWithRefEntries(t, refEntries)
}

func TestHashMapSmallRandom(t *testing.T) {
	for p := 0; p < SmallRandomPass; p++ {
		var refEntries []refEntry
		add := func(k testKey, v string) {
			refEntries = append(refEntries, refEntry{k, v})
		}

		for i := 0; i < NSmallRandom; i++ {
			k := uint64(uint64(rand.Int31n(SmallRandomHighBound))<<32 |
				uint64(rand.Int31n(SmallRandomLowBound)))
			add(testKey(k), "random "+hex(k))
		}

		testHashMapWithRefEntries(t, refEntries)
	}
}

func testHashMapWithRefEntries(t *testing.T, refEntries []refEntry) {
	m := Empty
	// Len of Empty should be 0.
	if m.Len() != 0 {
		t.Errorf("m.Len = %d, want %d", m.Len(), 0)
	}

	// Assoc and Len, test by building a map simutaneously.
	ref := make(map[testKey]string, len(refEntries))
	for _, e := range refEntries {
		ref[e.k] = e.v
		m = m.Assoc(e.k, e.v)
		if m.Len() != len(ref) {
			t.Errorf("m.Len = %d, want %d", m.Len(), len(ref))
		}
	}

	// Get.
	testMapContent(t, m, ref)
	in, got := m.Get(anotherTestKey(0))
	if in {
		t.Errorf("m.Get <bad key> returns entry %v", got)
	}
	// Iterator.
	testIterator(t, m, ref)

	// Without.
	// Ineffective ones.
	for i := 0; i < NIneffectiveWithout; i++ {
		k := anotherTestKey(uint32(rand.Int31())>>15 | uint32(rand.Int31())<<16)
		m = m.Without(k)
		if m.Len() != len(ref) {
			t.Errorf("m.Without removes item when it shouldn't")
		}
	}

	// Effective ones.
	for i := len(refEntries) - 1; i >= 0; i-- {
		k := refEntries[i].k
		delete(ref, k)
		m = m.Without(k)
		if m.Len() != len(ref) {
			t.Errorf("m.Len() = %d after removing, should be %v", m.Len(), len(ref))
		}
		in, _ := m.Get(k)
		if in {
			t.Errorf("m.Get(%v) still returns item after removal", k)
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

func testIterator(t *testing.T, m HashMap, ref map[testKey]string) {
	ref2 := map[Key]interface{}{}
	for k, v := range ref {
		ref2[k] = v
	}
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		if ref2[k] != v {
			t.Errorf("iterator yields unexpected pair %v, %v", k, v)
		}
		delete(ref2, k)
	}
	if len(ref2) != 0 {
		t.Errorf("iterating was not exhaustive")
	}
}
