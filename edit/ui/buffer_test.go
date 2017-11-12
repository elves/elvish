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
	{NewBuffer(10), Pos{0, 0}},
	{NewBuffer(10).SetLines([]Cell{{"a", 1, ""}}, []Cell{{"好", 2, ""}}),
		Pos{1, 2}},
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
	{[]*Buffer{
		NewBuffer(10).SetLines([]Cell{}, []Cell{}),
		NewBuffer(10),
		NewBuffer(10).SetLines([]Cell{}, []Cell{})},
		5},
}

func TestBuffersHeight(t *testing.T) {
	for _, test := range buffersHeighTests {
		if h := BuffersHeight(test.buffers...); h != test.want {
			t.Errorf("buffersHeight(%v...) = %v, want %v",
				test.buffers, h, test.want)
		}
	}
}

var bufferWritesTests = []struct {
	buf   *Buffer
	text  string
	style string
	want  *Buffer
}{
	// Writing nothing.
	{NewBuffer(10), "", "", NewBuffer(10)},
	// Writing a single rune.
	{NewBuffer(10), "a", "1", NewBuffer(10).SetLines([]Cell{{"a", 1, "1"}})},
	// Writing control character.
	{NewBuffer(10), "\033", "",
		NewBuffer(10).SetLines(
			[]Cell{{"^[", 2, styleForControlChar.String()}},
		)},
	// Writing styled control character.
	{NewBuffer(10), "a\033b", "1",
		NewBuffer(10).SetLines(
			[]Cell{
				{"a", 1, "1"},
				{"^[", 2, "1;" + styleForControlChar.String()},
				{"b", 1, "1"},
			},
		)},
	// Writing text containing a newline.
	{NewBuffer(10), "a\nb", "1",
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, "1"}}, []Cell{{"b", 1, "1"}},
		)},
	// Writing text containing a newline when there is indent.
	{NewBuffer(10).SetIndent(2), "a\nb", "1",
		NewBuffer(10).SetIndent(2).SetLines(
			[]Cell{{"a", 1, "1"}},
			[]Cell{{" ", 1, ""}, {" ", 1, ""}, {"b", 1, "1"}},
		)},
	// Writing long text that triggers wrapping.
	{NewBuffer(4), "aaaab", "1",
		NewBuffer(4).SetLines(
			[]Cell{{"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}},
			[]Cell{{"b", 1, "1"}},
		)},
	// Writing long text that triggers wrapping when there is indent.
	{NewBuffer(4).SetIndent(2), "aaaab", "1",
		NewBuffer(4).SetIndent(2).SetLines(
			[]Cell{{"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}},
			[]Cell{{" ", 1, ""}, {" ", 1, ""}, {"b", 1, "1"}},
		)},
	// Writing long text that triggers eager wrapping.
	{NewBuffer(4).SetIndent(2).SetEagerWrap(true), "aaaa", "1",
		NewBuffer(4).SetIndent(2).SetEagerWrap(true).SetLines(
			[]Cell{{"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}},
			[]Cell{{" ", 1, ""}, {" ", 1, ""}},
		)},
}

// TestBufferWrites tests buffer.writes by calling writes on a buffer and see if
// the buffer matches what is expected.
func TestBufferWrites(t *testing.T) {
	for _, test := range bufferWritesTests {
		b := test.buf
		b.WriteString(test.text, test.style)
		if !reflect.DeepEqual(b, test.want) {
			t.Errorf("buf.writes(%q, %q) makes it %v, want %v",
				test.text, test.style, b, test.want)
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
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}},
			[]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}},
		), 0, 2,
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}},
		),
	},
	// With dot.
	{
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}},
			[]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}},
		).SetDot(Pos{1, 1}), 1, 3,
		NewBuffer(10).SetLines(
			[]Cell{{"b", 1, ""}}, []Cell{{"c", 1, ""}},
		).SetDot(Pos{0, 1}),
	},
	// With dot that is going to be trimmed away.
	{
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}},
			[]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}},
		).SetDot(Pos{0, 1}), 1, 3,
		NewBuffer(10).SetLines(
			[]Cell{{"b", 1, ""}}, []Cell{{"c", 1, ""}},
		).SetDot(Pos{0, 1}),
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
		NewBuffer(10).SetLines([]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}}),
		NewBuffer(11).SetLines([]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}}),
		false,
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}},
			[]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}},
		),
	},
	// Moving dot.
	{
		NewBuffer(10).SetLines([]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}}),
		NewBuffer(11).SetLines(
			[]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}},
		).SetDot(Pos{1, 1}),
		true,
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}},
			[]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}},
		).SetDot(Pos{3, 1}),
	},
}

func TestExtend(t *testing.T) {
	for _, test := range bufferExtendTests {
		b := test.buf
		b.Extend(test.buf2, test.moveDot)
		if !reflect.DeepEqual(b, test.want) {
			t.Errorf("buf.extend(%v, %v) makes it %v, want %v",
				test.buf2, test.moveDot, b, test.want)
		}
	}
}

var bufferExtendRightTests = []struct {
	buf  *Buffer
	buf2 *Buffer
	w    int
	want *Buffer
}{
	// No padding, equal height.
	{
		NewBuffer(10).SetLines([]Cell{{"a", 1, ""}}, []Cell{}),
		NewBuffer(11).SetLines([]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}}),
		0,
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}, {"c", 1, ""}},
			[]Cell{{"d", 1, ""}},
		),
	},
	// With padding.
	{
		NewBuffer(10).SetLines([]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}}),
		NewBuffer(11).SetLines([]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}}),
		2,
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}, {" ", 1, ""}, {"c", 1, ""}},
			[]Cell{{"b", 1, ""}, {" ", 1, ""}, {"d", 1, ""}},
		),
	},
	// buf is higher.
	{
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}},
			[]Cell{{"b", 1, ""}},
			[]Cell{{"x", 1, ""}},
		),
		NewBuffer(11).SetLines([]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}}),
		1,
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}, {"c", 1, ""}},
			[]Cell{{"b", 1, ""}, {"d", 1, ""}},
			[]Cell{{"x", 1, ""}},
		),
	},
	// buf2 is higher.
	{
		NewBuffer(10).SetLines([]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}}),
		NewBuffer(11).SetLines(
			[]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}}, []Cell{{"e", 1, ""}},
		),
		1,
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}, {"c", 1, ""}},
			[]Cell{{"b", 1, ""}, {"d", 1, ""}},
			[]Cell{{" ", 1, ""}, {"e", 1, ""}},
		),
	},
}

func TestExtendRight(t *testing.T) {
	for _, test := range bufferExtendRightTests {
		b := test.buf
		b.ExtendRight(test.buf2, test.w)
		if !reflect.DeepEqual(b, test.want) {
			t.Errorf("buf.extendRight(%v, %v) makes it %v, want %v",
				test.buf2, test.w, b, test.want)
		}
	}
}
