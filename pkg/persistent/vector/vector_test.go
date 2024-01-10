package vector

import (
	"errors"
	"math/rand"
	"strconv"
	"testing"
)

// Nx is the minimum number of elements for the internal tree of the vector to
// be x levels deep.
const (
	N1 = tailMaxLen + 1                              // 33
	N2 = nodeSize + tailMaxLen + 1                   // 65
	N3 = nodeSize*nodeSize + tailMaxLen + 1          // 1057
	N4 = nodeSize*nodeSize*nodeSize + tailMaxLen + 1 // 32801
)

func TestVector(t *testing.T) {
	run := func(n int) {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			v := testConj(t, n)
			testIndex(t, v, 0, n)
			testAssoc(t, v, "233")
			testIterator(t, v.Iterator(), 0, n)
			testPop(t, v)
		})
	}

	for i := 0; i <= N3; i++ {
		run(i)
	}
	run(N4)
}

// Regression test against #4.
func TestIterator_VectorWithNil(t *testing.T) {
	run := func(n int) {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			v := Empty
			for i := 0; i < n; i++ {
				v = v.Conj(nil)
			}

			iterated := 0
			for it := v.Iterator(); it.HasElem(); it.Next() {
				iterated++
				if it.Elem() != nil {
					t.Errorf("element not nil")
				}
			}
			if iterated != n {
				t.Errorf("did not iterate %d items", n)
			}
		})
	}
	for i := 0; i <= N3; i++ {
		run(i)
	}
	run(N4)
}

// testConj creates a vector containing 0...n-1 with Conj, and ensures that the
// length of the old and new vectors are expected after each Conj. It returns
// the created vector.
func testConj(t *testing.T, n int) Vector {
	v := Empty
	for i := 0; i < n; i++ {
		oldv := v
		v = v.Conj(i)

		if count := oldv.Len(); count != i {
			t.Errorf("oldv.Count() == %v, want %v", count, i)
		}
		if count := v.Len(); count != i+1 {
			t.Errorf("v.Count() == %v, want %v", count, i+1)
		}
	}
	return v
}

// testIndex tests Index, assuming that the vector contains begin...int-1.
func testIndex(t *testing.T, v Vector, begin, end int) {
	n := v.Len()
	for i := 0; i < n; i++ {
		elem, _ := v.Index(i)
		if elem != i {
			t.Errorf("v.Index(%v) == %v, want %v", i, elem, i)
		}
	}
	for _, i := range []int{-2, -1, n, n + 1, n * 2} {
		if elem, _ := v.Index(i); elem != nil {
			t.Errorf("v.Index(%d) == %v, want nil", i, elem)
		}
	}
}

// testIterator tests the iterator, assuming that the result is begin...end-1.
func testIterator(t *testing.T, it Iterator, begin, end int) {
	i := begin
	for ; it.HasElem(); it.Next() {
		elem := it.Elem()
		if elem != i {
			t.Errorf("iterator produce %v, want %v", elem, i)
		}
		i++
	}
	if i != end {
		t.Errorf("iterator produces up to %v, want %v", i, end)
	}
}

// testAssoc tests Assoc by replacing each element.
func testAssoc(t *testing.T, v Vector, subst any) {
	n := v.Len()
	for i := 0; i <= n; i++ {
		oldv := v
		v = v.Assoc(i, subst)

		if i < n {
			elem, _ := oldv.Index(i)
			if elem != i {
				t.Errorf("oldv.Index(%v) == %v, want %v", i, elem, i)
			}
		}

		elem, _ := v.Index(i)
		if elem != subst {
			t.Errorf("v.Index(%v) == %v, want %v", i, elem, subst)
		}
	}

	n++
	for _, i := range []int{-1, n + 1, n + 2, n * 2} {
		newv := v.Assoc(i, subst)
		if newv != nil {
			t.Errorf("v.Assoc(%d) = %v, want nil", i, newv)
		}
	}
}

// testPop tests Pop by removing each element.
func testPop(t *testing.T, v Vector) {
	n := v.Len()
	for i := 0; i < n; i++ {
		oldv := v
		v = v.Pop()

		if count := oldv.Len(); count != n-i {
			t.Errorf("oldv.Count() == %v, want %v", count, n-i)
		}
		if count := v.Len(); count != n-i-1 {
			t.Errorf("oldv.Count() == %v, want %v", count, n-i-1)
		}
	}
	newv := v.Pop()
	if newv != nil {
		t.Errorf("v.Pop() = %v, want nil", newv)
	}
}

func TestSubVector(t *testing.T) {
	v := Empty
	for i := 0; i < 10; i++ {
		v = v.Conj(i)
	}

	sv := v.SubVector(0, 4)
	testIndex(t, sv, 0, 4)
	testAssoc(t, sv, "233")
	testIterator(t, sv.Iterator(), 0, 4)
	testPop(t, sv)

	sv = v.SubVector(1, 4)
	if !checkVector(sv, 1, 2, 3) {
		t.Errorf("v[0:4] is not expected")
	}
	if !checkVector(sv.Assoc(1, "233"), 1, "233", 3) {
		t.Errorf("v[0:4].Assoc is not expected")
	}
	if !checkVector(sv.Conj("233"), 1, 2, 3, "233") {
		t.Errorf("v[0:4].Conj is not expected")
	}
	if !checkVector(sv.Pop(), 1, 2) {
		t.Errorf("v[0:4].Pop is not expected")
	}
	if !checkVector(sv.SubVector(1, 2), 2) {
		t.Errorf("v[0:4][1:2] is not expected")
	}
	testIterator(t, sv.Iterator(), 1, 4)

	if !checkVector(v.SubVector(1, 1)) {
		t.Errorf("v[1:1] is not expected")
	}
	// Begin is allowed to be equal to n if end is also n
	if !checkVector(v.SubVector(10, 10)) {
		t.Errorf("v[10:10] is not expected")
	}

	bad := v.SubVector(-1, 0)
	if bad != nil {
		t.Errorf("v.SubVector(-1, 0) = %v, want nil", bad)
	}
	bad = v.SubVector(5, 100)
	if bad != nil {
		t.Errorf("v.SubVector(5, 100) = %v, want nil", bad)
	}
	bad = v.SubVector(-1, 100)
	if bad != nil {
		t.Errorf("v.SubVector(-1, 100) = %v, want nil", bad)
	}
	bad = v.SubVector(4, 2)
	if bad != nil {
		t.Errorf("v.SubVector(4, 2) = %v, want nil", bad)
	}
}

// Regression test for https://b.elv.sh/1287: crash when tree has a height >= 1
// and start of subvector is in the tail.
func TestSubVector_BeginFromTail(t *testing.T) {
	v := Empty
	for i := 0; i < 65; i++ {
		v = v.Conj(i)
	}
	sv := v.SubVector(64, 65)
	testIterator(t, sv.Iterator(), 64, 65)
}

func checkVector(v Vector, values ...any) bool {
	if v.Len() != len(values) {
		return false
	}
	for i, a := range values {
		if x, _ := v.Index(i); x != a {
			return false
		}
	}
	return true
}

func TestVectorEqual(t *testing.T) {
	v1, v2 := Empty, Empty
	for i := 0; i < N3; i++ {
		elem := rand.Int63()
		v1 = v1.Conj(elem)
		v2 = v2.Conj(elem)
		if !eqVector(v1, v2) {
			t.Errorf("Not equal after Conj'ing %d elements", i+1)
		}
	}
}

func eqVector(v1, v2 Vector) bool {
	if v1.Len() != v2.Len() {
		return false
	}
	for i := 0; i < v1.Len(); i++ {
		a1, _ := v1.Index(i)
		a2, _ := v2.Index(i)
		if a1 != a2 {
			return false
		}
	}
	return true
}

var marshalJSONTests = []struct {
	in      Vector
	wantOut string
	wantErr error
}{
	{makeVector("1", 2, nil), `["1",2,null]`, nil},
	{makeVector("1", makeVector(2)), `["1",[2]]`, nil},
	{makeVector(0, 1, 2, 3, 4, 5).SubVector(1, 5), `[1,2,3,4]`, nil},
	{makeVector(0, func() {}), "", errors.New("element 1: json: unsupported type: func()")},
}

func TestMarshalJSON(t *testing.T) {
	for i, test := range marshalJSONTests {
		out, err := test.in.MarshalJSON()
		if string(out) != test.wantOut {
			t.Errorf("v%d.MarshalJSON -> out %q, want %q", i, out, test.wantOut)
		}
		if err == nil || test.wantErr == nil {
			if err != test.wantErr {
				t.Errorf("v%d.MarshalJSON -> err %v, want %v", i, err, test.wantErr)
			}
		} else {
			if err.Error() != test.wantErr.Error() {
				t.Errorf("v%d.MarshalJSON -> err %v, want %v", i, err, test.wantErr)
			}
		}
	}
}

func makeVector(elements ...any) Vector {
	v := Empty
	for _, element := range elements {
		v = v.Conj(element)
	}
	return v
}

func BenchmarkConjNativeN1(b *testing.B) { benchmarkNativeAppend(b, N1) }
func BenchmarkConjNativeN2(b *testing.B) { benchmarkNativeAppend(b, N2) }
func BenchmarkConjNativeN3(b *testing.B) { benchmarkNativeAppend(b, N3) }
func BenchmarkConjNativeN4(b *testing.B) { benchmarkNativeAppend(b, N4) }

func benchmarkNativeAppend(b *testing.B, n int) {
	for r := 0; r < b.N; r++ {
		var s []any
		for i := 0; i < n; i++ {
			s = append(s, i)
		}
		_ = s
	}
}

func BenchmarkConjPersistentN1(b *testing.B) { benchmarkConj(b, N1) }
func BenchmarkConjPersistentN2(b *testing.B) { benchmarkConj(b, N2) }
func BenchmarkConjPersistentN3(b *testing.B) { benchmarkConj(b, N3) }
func BenchmarkConjPersistentN4(b *testing.B) { benchmarkConj(b, N4) }

func benchmarkConj(b *testing.B, n int) {
	for r := 0; r < b.N; r++ {
		v := Empty
		for i := 0; i < n; i++ {
			v = v.Conj(i)
		}
	}
}

var (
	sliceN4  = make([]any, N4)
	vectorN4 = Empty
)

func init() {
	for i := 0; i < N4; i++ {
		vectorN4 = vectorN4.Conj(i)
	}
}

var x any

func BenchmarkIndexSeqNativeN4(b *testing.B) { benchmarkIndexSeqNative(b, N4) }

func benchmarkIndexSeqNative(b *testing.B, n int) {
	for r := 0; r < b.N; r++ {
		for i := 0; i < n; i++ {
			x = sliceN4[i]
		}
	}
}

func BenchmarkIndexSeqPersistentN4(b *testing.B) { benchmarkIndexSeqPersistent(b, N4) }

func benchmarkIndexSeqPersistent(b *testing.B, n int) {
	for r := 0; r < b.N; r++ {
		for i := 0; i < n; i++ {
			x, _ = vectorN4.Index(i)
		}
	}
}

var randIndices []int

func init() {
	randIndices = make([]int, N4)
	for i := 0; i < N4; i++ {
		randIndices[i] = rand.Intn(N4)
	}
}

func BenchmarkIndexRandNative(b *testing.B) {
	for r := 0; r < b.N; r++ {
		for _, i := range randIndices {
			x = sliceN4[i]
		}
	}
}

func BenchmarkIndexRandPersistent(b *testing.B) {
	for r := 0; r < b.N; r++ {
		for _, i := range randIndices {
			x, _ = vectorN4.Index(i)
		}
	}
}

func nativeEqual(s1, s2 []int) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v1 := range s1 {
		if v1 != s2[i] {
			return false
		}
	}
	return true
}

func BenchmarkEqualNative(b *testing.B) {
	b.StopTimer()
	var s1, s2 []int
	for i := 0; i < N4; i++ {
		s1 = append(s1, i)
		s2 = append(s2, i)
	}
	b.StartTimer()

	for r := 0; r < b.N; r++ {
		eq := nativeEqual(s1, s2)
		if !eq {
			panic("not equal")
		}
	}
}

func BenchmarkEqualPersistent(b *testing.B) {
	b.StopTimer()
	v1, v2 := Empty, Empty
	for i := 0; i < N4; i++ {
		v1 = v1.Conj(i)
		v2 = v2.Conj(i)
	}
	b.StartTimer()

	for r := 0; r < b.N; r++ {
		eq := eqVector(v1, v2)
		if !eq {
			panic("not equal")
		}
	}
}
