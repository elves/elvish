package ui

import (
	"reflect"
	"testing"
)

var cellsWidthTests = []struct {
	cells     []Cell
	wantWidth int
}{
	{[]Cell{}, 0},
	{[]Cell{{"a", 1, ""}, {"好", 2, ""}}, 3},
}

func TestCellsWidth(t *testing.T) {
	for _, test := range cellsWidthTests {
		if width := CellsWidth(test.cells); width != test.wantWidth {
			t.Errorf("cellsWidth(%v) = %v, want %v",
				test.cells, width, test.wantWidth)
		}
	}
}

var makeSpacingTests = []struct {
	n    int
	want []Cell
}{
	{0, []Cell{}},
	{1, []Cell{{" ", 1, ""}}},
	{4, []Cell{{" ", 1, ""}, {" ", 1, ""}, {" ", 1, ""}, {" ", 1, ""}}},
}

func TestMakeSpacing(t *testing.T) {
	for _, test := range makeSpacingTests {
		if got := makeSpacing(test.n); !reflect.DeepEqual(got, test.want) {
			t.Errorf("makeSpacing(%v) = %v, want %v", test.n, got, test.want)
		}
	}
}

var compareCellsTests = []struct {
	cells1    []Cell
	cells2    []Cell
	wantEqual bool
	wantIndex int
}{
	{[]Cell{}, []Cell{}, true, 0},
	{[]Cell{}, []Cell{{"a", 1, ""}}, false, 0},
	{
		[]Cell{{"a", 1, ""}, {"好", 2, ""}, {"b", 1, ""}},
		[]Cell{{"a", 1, ""}, {"好", 2, ""}, {"c", 1, ""}},
		false, 2,
	},
	{
		[]Cell{{"a", 1, ""}, {"好", 2, ""}, {"b", 1, ""}},
		[]Cell{{"a", 1, ""}, {"好", 2, "1"}, {"c", 1, ""}},
		false, 1,
	},
}

func TestCompareCells(t *testing.T) {
	for _, test := range compareCellsTests {
		equal, index := CompareCells(test.cells1, test.cells2)
		if equal != test.wantEqual || index != test.wantIndex {
			t.Errorf("compareCells(%v, %v) = (%v, %v), want (%v, %v)",
				test.cells1, test.cells2,
				equal, index, test.wantEqual, test.wantIndex)
		}
	}
}

var bufferCursorTests = []struct {
	buf  *Buffer
	want Pos
}{
	{
		&Buffer{Width: 10, Lines: Lines{Line{}}},
		Pos{0, 0},
	},
	{
		&Buffer{Width: 10, Lines: Lines{Line{C("a", "")}, Line{C("好", "")}}},
		Pos{1, 2},
	},
}

func TestBufferCursor(t *testing.T) {
	for _, test := range bufferCursorTests {
		if p := test.buf.Cursor(); p != test.want {
			t.Errorf("(%v).cursor() = %v, want %v", test.buf, p, test.want)
		}
	}
}

var buffersHeighTests = []struct {
	buffers []*Buffer
	want    int
}{
	{nil, 0},
	{[]*Buffer{NewBuffer(10)}, 1},
	{
		[]*Buffer{
			&Buffer{Width: 10, Lines: Lines{Line{}, Line{}}},
			&Buffer{Width: 10, Lines: Lines{Line{}}},
			&Buffer{Width: 10, Lines: Lines{Line{}, Line{}}},
		},
		5,
	},
}

func TestBuffersHeight(t *testing.T) {
	for _, test := range buffersHeighTests {
		if h := BuffersHeight(test.buffers...); h != test.want {
			t.Errorf("buffersHeight(%v...) = %v, want %v",
				test.buffers, h, test.want)
		}
	}
}

var bufferTrimToLinesTests = []struct {
	buf  *Buffer
	low  int
	high int
	want *Buffer
}{
	{
		&Buffer{Width: 10, Lines: Lines{
			Line{C("a", "")}, Line{C("b", "")}, Line{C("c", "")}, Line{C("d", "")},
		}},
		0, 2,
		&Buffer{Width: 10, Lines: Lines{
			Line{C("a", "")}, Line{C("b", "")},
		}},
	},
	// With dot.
	{
		&Buffer{Width: 10, Lines: Lines{
			Line{C("a", "")}, Line{C("b", "")}, Line{C("c", "")}, Line{C("d", "")},
		}, Dot: Pos{1, 1}},
		1, 3,
		&Buffer{Width: 10, Lines: Lines{
			Line{C("b", "")}, Line{C("c", "")},
		}, Dot: Pos{0, 1}},
	},
	// With dot that is going to be trimmed away.
	{
		&Buffer{Width: 10, Lines: Lines{
			Line{C("a", "")}, Line{C("b", "")}, Line{C("c", "")}, Line{C("d", "")},
		}, Dot: Pos{0, 1}},
		1, 3,
		&Buffer{Width: 10, Lines: Lines{
			Line{C("b", "")}, Line{C("c", "")},
		}, Dot: Pos{0, 1}},
	},
}

func TestBufferTrimToLines(t *testing.T) {
	for _, test := range bufferTrimToLinesTests {
		b := test.buf
		b.TrimToLines(test.low, test.high)
		if !reflect.DeepEqual(b, test.want) {
			t.Errorf("buf.trimToLines(%v, %v) makes it %v, want %v",
				test.low, test.high, b, test.want)
		}
	}
}
