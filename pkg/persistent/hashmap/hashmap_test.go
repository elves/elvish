package hashmap

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	"src.elv.sh/pkg/persistent/hash"
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

	NIneffectiveDissoc = 0x200

	N1 = nodeCap + 1
	N2 = nodeCap*nodeCap + 1
	N3 = nodeCap*nodeCap*nodeCap + 1
)

type testKey uint64
type anotherTestKey uint32

func equalFunc(k1, k2 any) bool {
	switch k1 := k1.(type) {
	case testKey:
		t2, ok := k2.(testKey)
		return ok && k1 == t2
	case anotherTestKey:
		return false
	default:
		return k1 == k2
	}
}

func hashFunc(k any) uint32 {
	switch k := k.(type) {
	case uint32:
		return k
	case string:
		return hash.String(k)
	case testKey:
		// Return the lower 32 bits for testKey. This is intended so that hash
		// collisions can be easily constructed.
		return uint32(k & 0xffffffff)
	case anotherTestKey:
		return uint32(k)
	default:
		return 0
	}
}

var empty = New(equalFunc, hashFunc)

type refEntry struct {
	k testKey
	v string
}

func hex(i uint64) string {
	return "0x" + strconv.FormatUint(i, 16)
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

var marshalJSONTests = []struct {
	in      Map
	wantOut string
	wantErr bool
}{
	{makeHashMap(uint32(1), "a", "2", "b"), `{"1":"a","2":"b"}`, false},
	// Invalid key type
	{makeHashMap([]any{}, "x"), "", true},
}

func TestMarshalJSON(t *testing.T) {
	for i, test := range marshalJSONTests {
		out, err := test.in.MarshalJSON()
		if string(out) != test.wantOut {
			t.Errorf("m%d.MarshalJSON -> out %s, want %s", i, out, test.wantOut)
		}
		if (err != nil) != test.wantErr {
			var wantErr string
			if test.wantErr {
				wantErr = "non-nil"
			} else {
				wantErr = "nil"
			}
			t.Errorf("m%d.MarshalJSON -> err %v, want %s", i, err, wantErr)
		}
	}
}

func makeHashMap(data ...any) Map {
	m := empty
	for i := 0; i+1 < len(data); i += 2 {
		k, v := data[i], data[i+1]
		m = m.Assoc(k, v)
	}
	return m
}

// testHashMapWithRefEntries tests the operations of a Map. It uses the supplied
// list of entries to build the map, and then test all its operations.
func testHashMapWithRefEntries(t *testing.T, refEntries []refEntry) {
	m := empty
	// Len of Empty should be 0.
	if m.Len() != 0 {
		t.Errorf("m.Len = %d, want %d", m.Len(), 0)
	}

	// Assoc and Len, test by building a map simultaneously.
	ref := make(map[testKey]string, len(refEntries))
	for _, e := range refEntries {
		ref[e.k] = e.v
		m = m.Assoc(e.k, e.v)
		if m.Len() != len(ref) {
			t.Errorf("m.Len = %d, want %d", m.Len(), len(ref))
		}
	}

	// Index.
	testMapContent(t, m, ref)
	got, in := m.Index(anotherTestKey(0))
	if in {
		t.Errorf("m.Index <bad key> returns entry %v", got)
	}
	// Iterator.
	testIterator(t, m, ref)

	// Dissoc.
	// Ineffective ones.
	for i := 0; i < NIneffectiveDissoc; i++ {
		k := anotherTestKey(uint32(rand.Int31())>>15 | uint32(rand.Int31())<<16)
		m = m.Dissoc(k)
		if m.Len() != len(ref) {
			t.Errorf("m.Dissoc removes item when it shouldn't")
		}
	}

	// Effective ones.
	for x := 0; x < len(refEntries); x++ {
		i := rand.Intn(len(refEntries))
		k := refEntries[i].k
		delete(ref, k)
		m = m.Dissoc(k)
		if m.Len() != len(ref) {
			t.Errorf("m.Len() = %d after removing, should be %v", m.Len(), len(ref))
		}
		_, in := m.Index(k)
		if in {
			t.Errorf("m.Index(%v) still returns item after removal", k)
		}
		// Checking all elements is expensive. Only do this 1% of the time.
		if rand.Float64() < 0.01 {
			testMapContent(t, m, ref)
			testIterator(t, m, ref)
		}
	}
}

func testMapContent(t *testing.T, m Map, ref map[testKey]string) {
	for k, v := range ref {
		got, in := m.Index(k)
		if !in {
			t.Errorf("m.Index 0x%x returns no entry", k)
		}
		if got != v {
			t.Errorf("m.Index(0x%x) = %v, want %v", k, got, v)
		}
	}
}

func testIterator(t *testing.T, m Map, ref map[testKey]string) {
	ref2 := map[any]any{}
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

func TestNilKey(t *testing.T) {
	m := empty

	testLen := func(l int) {
		if m.Len() != l {
			t.Errorf(".Len -> %d, want %d", m.Len(), l)
		}
	}
	testIndex := func(wantVal any, wantOk bool) {
		val, ok := m.Index(nil)
		if val != wantVal {
			t.Errorf(".Index -> %v, want %v", val, wantVal)
		}
		if ok != wantOk {
			t.Errorf(".Index -> ok %v, want %v", ok, wantOk)
		}
	}

	testLen(0)
	testIndex(nil, false)

	m = m.Assoc(nil, "nil value")
	testLen(1)
	testIndex("nil value", true)

	m = m.Assoc(nil, "nil value 2")
	testLen(1)
	testIndex("nil value 2", true)

	m = m.Dissoc(nil)
	testLen(0)
	testIndex(nil, false)
}

func TestIterateMapWithNilKey(t *testing.T) {
	m := empty.Assoc("k", "v").Assoc(nil, "nil value")
	var collected []any
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		collected = append(collected, k, v)
	}
	wantCollected := []any{nil, "nil value", "k", "v"}
	if !reflect.DeepEqual(collected, wantCollected) {
		t.Errorf("collected %v, want %v", collected, wantCollected)
	}
}

func BenchmarkSequentialConjNative1(b *testing.B) { nativeSequentialAdd(b.N, N1) }
func BenchmarkSequentialConjNative2(b *testing.B) { nativeSequentialAdd(b.N, N2) }
func BenchmarkSequentialConjNative3(b *testing.B) { nativeSequentialAdd(b.N, N3) }

// nativeSequentialAdd starts with an empty native map and adds elements 0...n-1
// to the map, using the same value as the key, repeating for N times.
func nativeSequentialAdd(N int, n uint32) {
	for r := 0; r < N; r++ {
		m := make(map[uint32]uint32)
		for i := uint32(0); i < n; i++ {
			m[i] = i
		}
	}
}

func BenchmarkSequentialConjPersistent1(b *testing.B) { sequentialConj(b.N, N1) }
func BenchmarkSequentialConjPersistent2(b *testing.B) { sequentialConj(b.N, N2) }
func BenchmarkSequentialConjPersistent3(b *testing.B) { sequentialConj(b.N, N3) }

// sequentialConj starts with an empty hash map and adds elements 0...n-1 to the
// map, using the same value as the key, repeating for N times.
func sequentialConj(N int, n uint32) {
	for r := 0; r < N; r++ {
		m := empty
		for i := uint32(0); i < n; i++ {
			m = m.Assoc(i, i)
		}
	}
}

func BenchmarkRandomStringsConjNative1(b *testing.B) { nativeRandomStringsAdd(b, N1) }
func BenchmarkRandomStringsConjNative2(b *testing.B) { nativeRandomStringsAdd(b, N2) }
func BenchmarkRandomStringsConjNative3(b *testing.B) { nativeRandomStringsAdd(b, N3) }

// nativeRandomStringsAdd starts with an empty native map and adds n random strings
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

func BenchmarkRandomStringsConjPersistent1(b *testing.B) { randomStringsConj(b, N1) }
func BenchmarkRandomStringsConjPersistent2(b *testing.B) { randomStringsConj(b, N2) }
func BenchmarkRandomStringsConjPersistent3(b *testing.B) { randomStringsConj(b, N3) }

func randomStringsConj(b *testing.B, n int) {
	ss := getRandomStrings(b)
	for r := 0; r < b.N; r++ {
		m := empty
		for i := 0; i < n; i++ {
			s := ss[i]
			m = m.Assoc(s, s)
		}
	}
}
