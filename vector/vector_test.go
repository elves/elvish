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

	for i := 0; i < n; i++ {
		elem := v.Nth(i)
		if num, ok := elem.(int); !ok || num != i {
			t.Errorf("v.Nth(%v) == %v, want %v", i, elem, i)
		}
	}

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
	sv := v.SubVector(0, 4)
	if !checkVector(sv, 0, 1, 2, 3) {
		t.Errorf("v[0:4] is not expected")
	}
	if !checkVector(sv.AssocN(1, "233"), 0, "233", 2, 3) {
		t.Errorf("v[0:4].AssocN is not expected")
	}
	if !checkVector(sv.Cons("233"), 0, 1, 2, 3, "233") {
		t.Errorf("v[0:4].Cons is not expected")
	}
	if !checkVector(sv.Pop(), 0, 1, 2) {
		t.Errorf("v[0:4].Pop is not expected")
	}
	if !checkVector(sv.SubVector(1, 3), 1, 2) {
		t.Errorf("v[0:4][1:2] is not expected")
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
