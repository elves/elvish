package edit

import (
	"reflect"
	"testing"
)

var cellsWidthTests = []struct {
	cells     []cell
	wantWidth int
}{
	{[]cell{}, 0},
	{[]cell{{"a", 1, ""}, {"好", 2, ""}}, 3},
}

func TestCellsWidth(t *testing.T) {
	for _, test := range cellsWidthTests {
		if width := cellsWidth(test.cells); width != test.wantWidth {
			t.Errorf("cellsWidth(%#v) = %#v, want %#v",
				test.cells, width, test.wantWidth)
		}
	}
}

var makeSpacingTests = []struct {
	n    int
	want []cell
}{
	{0, []cell{}},
	{1, []cell{{" ", 1, ""}}},
	{4, []cell{{" ", 1, ""}, {" ", 1, ""}, {" ", 1, ""}, {" ", 1, ""}}},
}

func TestMakeSpacing(t *testing.T) {
	for _, test := range makeSpacingTests {
		if got := makeSpacing(test.n); !reflect.DeepEqual(got, test.want) {
			t.Errorf("makeSpacing(%v) = %#v, want %#v", test.n, got, test.want)
		}
	}
}

var compareCellsTests = []struct {
	cells1    []cell
	cells2    []cell
	wantEqual bool
	wantIndex int
}{
	{[]cell{}, []cell{}, true, 0},
	{[]cell{}, []cell{{"a", 1, ""}}, false, 0},
	{
		[]cell{{"a", 1, ""}, {"好", 2, ""}, {"b", 1, ""}},
		[]cell{{"a", 1, ""}, {"好", 2, ""}, {"c", 1, ""}},
		false, 2,
	},
	{
		[]cell{{"a", 1, ""}, {"好", 2, ""}, {"b", 1, ""}},
		[]cell{{"a", 1, ""}, {"好", 2, "1"}, {"c", 1, ""}},
		false, 1,
	},
}

func TestCompareCells(t *testing.T) {
	for _, test := range compareCellsTests {
		equal, index := compareCells(test.cells1, test.cells2)
		if equal != test.wantEqual || index != test.wantIndex {
			t.Errorf("compareCells(%#v, %#v) = (%#v, %#v), want (%#v, %#v)",
				test.cells1, test.cells2,
				equal, index, test.wantEqual, test.wantIndex)
		}
	}
}
