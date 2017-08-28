package vector

import (
	"math/rand"
	"testing"
	"time"
)

// Nx is the minimum number of elements for the internal tree of the vector to
// be x levels deep.
const (
	N1 = tailMaxLen + 1                              // 33
	N2 = nodeSize + tailMaxLen + 1                   // 65
	N3 = nodeSize*nodeSize + tailMaxLen + 1          // 1057
	N4 = nodeSize*nodeSize*nodeSize + tailMaxLen + 1 // 32801
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func TestVector(t *testing.T) {
	const (
		subst = "233"
		n     = N4
	)

	v := testCons(t, n)
	testNth(t, v, 0, n)
	testAssocN(t, v, subst)
	testIterator(t, v.Iterator(), 0, n)
	testPop(t, v)
}

// testCons creates a vector containing 0...n-1 with Cons, and ensures that the
// length of the old and new vectors are expected after each Cons. It returns
// the created vector.
func testCons(t *testing.T, n int) Vector {
	v := Empty
	for i := 0; i < n; i++ {
		oldv := v
		v = v.Cons(i)

		if count := oldv.Len(); count != i {
			t.Errorf("oldv.Count() == %v, want %v", count, i)
		}
		if count := v.Len(); count != i+1 {
			t.Errorf("v.Count() == %v, want %v", count, i+1)
		}
	}
	return v
}

// testNth tests Nth, assuming that the vector contains begin...int-1.
func testNth(t *testing.T, v Vector, begin, end int) {
	n := v.Len()
	for i := 0; i < n; i++ {
		elem := v.Nth(i)
		if elem != i {
			t.Errorf("v.Nth(%v) == %v, want %v", i, elem, i)
		}
	}
	for _, i := range []int{-2, -1, n, n + 1, n * 2} {
		if elem := v.Nth(i); elem != nil {
			t.Errorf("v.Nth(%d) == %v, want nil", i, elem)
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

// testAssocN tests AssocN by replacing each element.
func testAssocN(t *testing.T, v Vector, subst interface{}) {
	n := v.Len()
	for i := 0; i <= n; i++ {
		oldv := v
		v = v.AssocN(i, subst)

		if i < n {
			elem := oldv.Nth(i)
			if elem != i {
				t.Errorf("oldv.Nth(%v) == %v, want %v", i, elem, i)
			}
		}

		elem := v.Nth(i)
		if elem != subst {
			t.Errorf("v.Nth(%v) == %v, want %v", i, elem, subst)
		}
	}

	n++
	for _, i := range []int{-1, n + 1, n + 2, n * 2} {
		newv := v.AssocN(i, subst)
		if newv != nil {
			t.Errorf("v.AssocN(%d) = %v, want nil", i, newv)
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
		v = v.Cons(i)
	}

	sv := v.SubVector(0, 4)
	testNth(t, sv, 0, 4)
	testAssocN(t, sv, "233")
	testIterator(t, sv.Iterator(), 0, 4)
	testPop(t, sv)

	sv = v.SubVector(1, 4)
	if !checkVector(sv, 1, 2, 3) {
		t.Errorf("v[0:4] is not expected")
	}
	if !checkVector(sv.AssocN(1, "233"), 1, "233", 3) {
		t.Errorf("v[0:4].AssocN is not expected")
	}
	if !checkVector(sv.Cons("233"), 1, 2, 3, "233") {
		t.Errorf("v[0:4].Cons is not expected")
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

func checkVector(v Vector, values ...interface{}) bool {
	if v.Len() != len(values) {
		return false
	}
	for i, a := range values {
		if v.Nth(i) != a {
			return false
		}
	}
	return true
}

func TestVectorEqual(t *testing.T) {
	v1, v2 := Empty, Empty
	for i := 0; i < N3; i++ {
		elem := rand.Int63()
		v1 = v1.Cons(elem)
		v2 = v2.Cons(elem)
		if !v1.Equal(v2) {
			t.Errorf("Not equal after Cons'ing %d elements", i+1)
		}
	}
}

func BenchmarkNativeAppendN1(b *testing.B) {
	benchmarkNativeAppend(b, N1)
}

func BenchmarkNativeAppendN2(b *testing.B) {
	benchmarkNativeAppend(b, N2)
}

func BenchmarkNativeAppendN3(b *testing.B) {
	benchmarkNativeAppend(b, N3)
}

func BenchmarkNativeAppendN4(b *testing.B) {
	benchmarkNativeAppend(b, N4)
}

func benchmarkNativeAppend(b *testing.B, n int) {
	for r := 0; r < b.N; r++ {
		var s []interface{}
		for i := 0; i < n; i++ {
			s = append(s, i)
		}
	}
}

func BenchmarkConsN1(b *testing.B) {
	benchmarkCons(b, N1)
}

func BenchmarkConsN2(b *testing.B) {
	benchmarkCons(b, N2)
}

func BenchmarkConsN3(b *testing.B) {
	benchmarkCons(b, N3)
}

func BenchmarkConsN4(b *testing.B) {
	benchmarkCons(b, N4)
}

func benchmarkCons(b *testing.B, n int) {
	for r := 0; r < b.N; r++ {
		v := Empty
		for i := 0; i < n; i++ {
			v = v.Cons(i)
		}
	}
}

var (
	sliceN4  = make([]interface{}, N4)
	vectorN4 = Empty
)

func init() {
	for i := 0; i < N4; i++ {
		vectorN4 = vectorN4.Cons(i)
	}
}

func BenchmarkNaitiveNth(b *testing.B) {
	for r := 0; r < b.N; r++ {
		for i := 0; i < N4; i++ {
			_ = sliceN4[i]
		}
	}
}

func BenchmarkNth(b *testing.B) {
	for r := 0; r < b.N; r++ {
		for i := 0; i < N4; i++ {
			_ = vectorN4.Nth(i)
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

func BenchmarkNativeEqual(b *testing.B) {
	b.StopTimer()
	var s1, s2 []int
	for i := 0; i < N4; i++ {
		s1 = append(s1, i)
		s2 = append(s2, i)
	}
	b.StartTimer()

	for r := 0; r < b.N; r++ {
		nativeEqual(s1, s2)
	}
}

func BenchmarkEqual(b *testing.B) {
	b.StopTimer()
	v1, v2 := Empty, Empty
	for i := 0; i < N4; i++ {
		v1 = v1.Cons(i)
		v2 = v2.Cons(i)
	}
	b.StartTimer()

	for r := 0; r < b.N; r++ {
		v1.Equal(v2)
	}
}
