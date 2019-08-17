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
	{[]Cell{{"a", ""}, {"好", ""}}, 3},
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
	{1, []Cell{{" ", ""}}},
	{4, []Cell{{" ", ""}, {" ", ""}, {" ", ""}, {" ", ""}}},
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
	{[]Cell{}, []Cell{{"a", ""}}, false, 0},
	{
		[]Cell{{"a", ""}, {"好", ""}, {"b", ""}},
		[]Cell{{"a", ""}, {"好", ""}, {"c", ""}},
		false, 2,
	},
	{
		[]Cell{{"a", ""}, {"好", ""}, {"b", ""}},
		[]Cell{{"a", ""}, {"好", "1"}, {"c", ""}},
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
		&Buffer{Width: 10, Lines: Lines{Line{Cell{"a", ""}}, Line{Cell{"好", ""}}}},
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
			{Width: 10, Lines: Lines{Line{}, Line{}}},
			{Width: 10, Lines: Lines{Line{}}},
			{Width: 10, Lines: Lines{Line{}, Line{}}},
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
			Line{Cell{"a", ""}}, Line{Cell{"b", ""}}, Line{Cell{"c", ""}}, Line{Cell{"d", ""}},
		}},
		0, 2,
		&Buffer{Width: 10, Lines: Lines{
			Line{Cell{"a", ""}}, Line{Cell{"b", ""}},
		}},
	},
	// With dot.
	{
		&Buffer{Width: 10, Lines: Lines{
			Line{Cell{"a", ""}}, Line{Cell{"b", ""}}, Line{Cell{"c", ""}}, Line{Cell{"d", ""}},
		}, Dot: Pos{1, 1}},
		1, 3,
		&Buffer{Width: 10, Lines: Lines{
			Line{Cell{"b", ""}}, Line{Cell{"c", ""}},
		}, Dot: Pos{0, 1}},
	},
	// With dot that is going to be trimmed away.
	{
		&Buffer{Width: 10, Lines: Lines{
			Line{Cell{"a", ""}}, Line{Cell{"b", ""}}, Line{Cell{"c", ""}}, Line{Cell{"d", ""}},
		}, Dot: Pos{0, 1}},
		1, 3,
		&Buffer{Width: 10, Lines: Lines{
			Line{Cell{"b", ""}}, Line{Cell{"c", ""}},
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

var bufferExtendTests = []struct {
	buf     *Buffer
	buf2    *Buffer
	moveDot bool
	want    *Buffer
}{
	{
		&Buffer{Width: 10, Lines: Lines{
			Line{Cell{"a", ""}}, Line{Cell{"b", ""}}}},
		&Buffer{Width: 11, Lines: Lines{
			Line{Cell{"c", ""}}, Line{Cell{"d", ""}}}},
		false,
		&Buffer{Width: 10, Lines: Lines{
			Line{Cell{"a", ""}}, Line{Cell{"b", ""}},
			Line{Cell{"c", ""}}, Line{Cell{"d", ""}}}},
	},
	// Moving dot.
	{
		&Buffer{Width: 10, Lines: Lines{
			Line{Cell{"a", ""}}, Line{Cell{"b", ""}}}},
		&Buffer{
			Width: 11,
			Lines: Lines{Line{Cell{"c", ""}}, Line{Cell{"d", ""}}},
			Dot:   Pos{1, 1},
		},
		true,
		&Buffer{
			Width: 10,
			Lines: Lines{
				Line{Cell{"a", ""}}, Line{Cell{"b", ""}},
				Line{Cell{"c", ""}}, Line{Cell{"d", ""}}},
			Dot: Pos{3, 1},
		},
	},
}

func TestBufferExtend(t *testing.T) {
	for _, test := range bufferExtendTests {
		buf := test.buf
		buf.Extend(test.buf2, test.moveDot)
		if !reflect.DeepEqual(buf, test.want) {
			t.Errorf("buf.extend(%v, %v) makes it %v, want %v",
				test.buf2, test.moveDot, buf, test.want)
		}
	}
}

var bufferExtendRightTests = []struct {
	buf  *Buffer
	buf2 *Buffer
	want *Buffer
}{
	// No padding, equal height.
	{
		&Buffer{Width: 1, Lines: Lines{Line{Cell{"a", ""}}, Line{Cell{"b", ""}}}},
		&Buffer{Width: 1, Lines: Lines{Line{Cell{"c", ""}}, Line{Cell{"d", ""}}}},
		&Buffer{Width: 2, Lines: Lines{
			Line{Cell{"a", ""}, Cell{"c", ""}},
			Line{Cell{"b", ""}, Cell{"d", ""}}}},
	},
	// With padding, equal height.
	{
		&Buffer{Width: 2, Lines: Lines{Line{Cell{"a", ""}}, Line{Cell{"b", ""}}}},
		&Buffer{Width: 1, Lines: Lines{Line{Cell{"c", ""}}, Line{Cell{"d", ""}}}},
		&Buffer{Width: 3, Lines: Lines{
			Line{Cell{"a", ""}, Cell{" ", ""}, Cell{"c", ""}},
			Line{Cell{"b", ""}, Cell{" ", ""}, Cell{"d", ""}}}},
	},
	// buf is higher.
	{
		&Buffer{Width: 1, Lines: Lines{
			Line{Cell{"a", ""}}, Line{Cell{"b", ""}}, Line{Cell{"x", ""}}}},
		&Buffer{Width: 1, Lines: Lines{
			Line{Cell{"c", ""}}, Line{Cell{"d", ""}},
		}},
		&Buffer{Width: 2, Lines: Lines{
			Line{Cell{"a", ""}, Cell{"c", ""}},
			Line{Cell{"b", ""}, Cell{"d", ""}},
			Line{Cell{"x", ""}}}},
	},
	// buf2 is higher.
	{
		&Buffer{Width: 1, Lines: Lines{Line{Cell{"a", ""}}, Line{Cell{"b", ""}}}},
		&Buffer{Width: 1, Lines: Lines{
			Line{Cell{"c", ""}}, Line{Cell{"d", ""}}, Line{Cell{"e", ""}},
		}},
		&Buffer{Width: 2, Lines: Lines{
			Line{Cell{"a", ""}, Cell{"c", ""}},
			Line{Cell{"b", ""}, Cell{"d", ""}},
			Line{Cell{" ", ""}, Cell{"e", ""}}}},
	},
}

func TestBufferExtendRight(t *testing.T) {
	for _, test := range bufferExtendRightTests {
		buf := test.buf
		buf.ExtendRight(test.buf2)
		if !reflect.DeepEqual(buf, test.want) {
			t.Errorf("buf.extendRight(%v) makes it %v, want %v",
				test.buf2, buf, test.want)
		}
	}
}
