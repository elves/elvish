package vector

import "testing"

// Nx is the minimum number of elements for the internal tree of the vector to
// be x levels deep.
const (
	N1 = tailMaxLen + 1                              // 33
	N2 = nodeSize + tailMaxLen + 1                   // 65
	N3 = nodeSize*nodeSize + tailMaxLen + 1          // 1057
	N4 = nodeSize*nodeSize*nodeSize + tailMaxLen + 1 // 32801
)

func TestVector(t *testing.T) {
	const (
		subst = "233"
		n     = N4
	)

	v := testCons(t, n)
	testNth(t, v)
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

// testNth tests Nth, assuming that the vector contains 0...n-1.
func testNth(t *testing.T, v Vector) {
	n := v.Len()
	for i := 0; i < n; i++ {
		elem := v.Nth(i)
		if num, ok := elem.(int); !ok || num != i {
			t.Errorf("v.Nth(%v) == %v, want %v", i, elem, i)
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
	for i := 0; i < n; i++ {
		oldv := v
		v = v.AssocN(i, subst)

		elem := oldv.Nth(i)
		if num, ok := elem.(int); !ok || num != i {
			t.Errorf("oldv.Nth(%v) == %v, want %v", i, elem, i)
		}

		elem = v.Nth(i)
		if str, ok := elem.(string); !ok || str != subst {
			t.Errorf("v.Nth(%v) == %v, want %v", i, elem, subst)
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
}

func TestSubVector(t *testing.T) {
	v := Empty
	for i := 0; i < 10; i++ {
		v = v.Cons(i)
	}
	sv := v.SubVector(1, 4)
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
