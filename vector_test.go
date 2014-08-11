package persistent

import "testing"

const (
	n = nodeCap*nodeCap + tailMaxLen + 1
)

func TestVector(t *testing.T) {
	const (
		subst = "233"
	)

	v := &Vector{}
	var i uint
	for i = 0; i < n; i++ {
		oldv := v
		v = v.Cons(i)

		if count := oldv.Count(); count != i {
			t.Errorf("oldv.Count() == %v, want %v", count, i)
		}
		if count := v.Count(); count != i+1 {
			t.Errorf("v.Count() == %v, want %v", count, i+1)
		}
	}

	for i = 0; i < n; i++ {
		elem := v.Nth(i)
		if num, ok := elem.(uint); !ok || num != i {
			t.Errorf("v.Nth(%v) == %v, want %v", i, elem, i)
		}
	}

	for i = 0; i < n; i++ {
		oldv := v
		v = v.AssocN(i, subst)

		elem := oldv.Nth(i)
		if num, ok := elem.(uint); !ok || num != i {
			t.Errorf("oldv.Nth(%v) == %v, want %v", i, elem, i)
		}

		elem = v.Nth(i)
		if str, ok := elem.(string); !ok || str != subst {
			t.Errorf("v.Nth(%v) == %v, want %v", i, elem, subst)
		}
	}

	for i = 0; i < n; i++ {
		oldv := v
		v = v.Pop()

		if count := oldv.Count(); count != n-i {
			t.Errorf("oldv.Count() == %v, want %v", count, n-i)
		}
		if count := v.Count(); count != n-i-1 {
			t.Errorf("oldv.Count() == %v, want %v", count, n-i-1)
		}
	}
}

func BenchmarkNativeAppend(b *testing.B) {
	for r := 0; r < b.N; r++ {
		var s []interface{}
		var i uint
		for i = 0; i < n; i++ {
			s = append(s, i)
		}
	}
}

func BenchmarkCons(b *testing.B) {
	for r := 0; r < b.N; r++ {
		v := &Vector{}
		var i uint
		for i = 0; i < n; i++ {
			v = v.Cons(i)
		}
	}
}
