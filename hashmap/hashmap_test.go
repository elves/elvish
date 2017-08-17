package hashmap

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

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

	N1 = nodeCap + 1
	N2 = nodeCap*nodeCap + 1
	N3 = nodeCap*nodeCap*nodeCap + 1
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

type refEntry struct {
	k testKey
	v string
}

func hex(i uint64) string {
	return "0x" + strconv.FormatUint(i, 16)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

var randomStrings []string

// getRandomStrings returns a slice of N3 random strings. It builds the slice
// once and caches it. If the slice is built for the first time, it stops the
// timer of the benchmark.
func getRandomStrings(b *testing.B) []string {
	if randomStrings == nil {
		b.StopTimer()
		defer b.StartTimer()
		randomStrings = make([]string, N3)
		for i := 0; i < N3; i++ {
			randomStrings[i] = makeRandomString()
		}
	}
	return randomStrings
}

// makeRandomString builds a random string consisting of n bytes (randomized
// between 0 and 99) and each byte is randomized between 0 and 255. The string
// need not be valid UTF-8.
func makeRandomString() string {
	bytes := make([]byte, rand.Intn(100))
	for i := range bytes {
		bytes[i] = byte(rand.Intn(256))
	}
	return string(bytes)
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

// testHashMapWithRefEntries tests the operations of a HashMap. It uses the
// supplied list of entries to build the hash map, and then test all its
// operations.
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

func BenchmarkNativeSequentialAddN1(b *testing.B) { nativeSequentialAdd(b.N, N1) }
func BenchmarkNativeSequentialAddN2(b *testing.B) { nativeSequentialAdd(b.N, N2) }
func BenchmarkNativeSequentialAddN3(b *testing.B) { nativeSequentialAdd(b.N, N3) }

// nativeSequntialAdd starts with an empty native map and adds elements 0...n-1
// to the map, using the same value as the key, repeating for N times.
func nativeSequentialAdd(N int, n uint32) {
	for r := 0; r < N; r++ {
		m := make(map[uint32]uint32)
		for i := uint32(0); i < n; i++ {
			m[i] = i
		}
	}
}

func BenchmarkSequentialConsN1(b *testing.B) { sequentialCons(b.N, N1) }
func BenchmarkSequentialConsN2(b *testing.B) { sequentialCons(b.N, N2) }
func BenchmarkSequentialConsN3(b *testing.B) { sequentialCons(b.N, N3) }

// sequentialCons starts with an empty HashMap and adds elements 0...n-1 to the
// map, using the same value as the key, repeating for N times.
func sequentialCons(N int, n UInt32) {
	for r := 0; r < N; r++ {
		m := Empty
		for i := UInt32(0); i < n; i++ {
			m = m.Assoc(i, i)
		}
	}
}

func BenchmarkNativeRandomStringsAddN1(b *testing.B) { nativeRandomStringsAdd(b, N1) }
func BenchmarkNativeRandomStringsAddN2(b *testing.B) { nativeRandomStringsAdd(b, N2) }
func BenchmarkNativeRandomStringsAddN3(b *testing.B) { nativeRandomStringsAdd(b, N3) }

// nativeSequntialAdd starts with an empty native map and adds n random strings
// to the map, using the same value as the key, repeating for b.N times.
func nativeRandomStringsAdd(b *testing.B, n int) {
	ss := getRandomStrings(b)
	for r := 0; r < b.N; r++ {
		m := make(map[string]string)
		for i := 0; i < n; i++ {
			s := ss[i]
			m[s] = s
		}
	}
}

func BenchmarkRandomStringsConsN1(b *testing.B) { randomStringsCons(b, N1) }
func BenchmarkRandomStringsConsN2(b *testing.B) { randomStringsCons(b, N2) }
func BenchmarkRandomStringsConsN3(b *testing.B) { randomStringsCons(b, N3) }

func randomStringsCons(b *testing.B, n int) {
	ss := getRandomStrings(b)
	for r := 0; r < b.N; r++ {
		m := Empty
		for i := 0; i < n; i++ {
			s := String(ss[i])
			m = m.Assoc(s, s)
		}
	}
}
