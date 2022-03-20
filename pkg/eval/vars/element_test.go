package vars

import (
	"testing"

	"src.elv.sh/pkg/eval/vals"
)

var elementTests = []struct {
	name         string
	oldContainer any
	indices      []any
	elemValue    any
	newContainer any
}{
	{
		"single level",
		vals.MakeMap("k1", "v1", "k2", "v2"),
		[]any{"k1"}, "new v1",
		vals.MakeMap("k1", "new v1", "k2", "v2"),
	},
	{
		"multi level",
		vals.MakeMap(
			"k1", vals.MakeMap("k1a", "v1a", "k1b", "v1b"), "k2", "v2"),
		[]any{"k1", "k1a"}, "new v1a",
		vals.MakeMap(
			"k1", vals.MakeMap("k1a", "new v1a", "k1b", "v1b"), "k2", "v2"),
	},
}

func TestElement(t *testing.T) {
	for _, test := range elementTests {
		t.Run(test.name, func(t *testing.T) {
			m := test.oldContainer

			elemVar, err := MakeElement(FromPtr(&m), test.indices)
			if err != nil {
				t.Errorf("MakeElement -> error %v, want nil", err)
			}

			elemVar.Set(test.elemValue)
			if !vals.Equal(m, test.newContainer) {
				t.Errorf("Value after Set is %v, want %v", m, test.newContainer)
			}

			if elemVar.Get() != test.elemValue {
				t.Errorf("elemVar.Get() -> %v, want %v",
					elemVar.Get(), test.elemValue)
			}
		})
	}
}

var delElementTests = []struct {
	name         string
	oldContainer any
	indices      []any
	newContainer any
}{
	{
		"single level",
		vals.MakeMap("k1", "v1", "k2", "v2"),
		[]any{"k1"},
		vals.MakeMap("k2", "v2"),
	},
	{
		"multi level",
		vals.MakeMap(
			"k1", vals.MakeMap("k1a", "v1a", "k1b", "v1b"), "k2", "v2"),
		[]any{"k1", "k1a"},
		vals.MakeMap("k1", vals.MakeMap("k1b", "v1b"), "k2", "v2"),
	},
}

func TestDelElement(t *testing.T) {
	for _, test := range delElementTests {
		t.Run(test.name, func(t *testing.T) {
			m := test.oldContainer

			DelElement(FromPtr(&m), test.indices)
			if !vals.Equal(m, test.newContainer) {
				t.Errorf("After deleting, map is %v, want %v",
					vals.ReprPlain(m),
					vals.ReprPlain(test.newContainer))
			}
		})
	}
}
