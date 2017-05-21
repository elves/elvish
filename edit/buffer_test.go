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
			t.Errorf("cellsWidth(%v) = %v, want %v",
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
			t.Errorf("makeSpacing(%v) = %v, want %v", test.n, got, test.want)
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
			t.Errorf("compareCells(%v, %v) = (%v, %v), want (%v, %v)",
				test.cells1, test.cells2,
				equal, index, test.wantEqual, test.wantIndex)
		}
	}
}

var bufferCursorTests = []struct {
	buf  *buffer
	want Pos
}{
	{newBuffer(10), Pos{0, 0}},
	{newBuffer(10).setLines([]cell{{"a", 1, ""}}, []cell{{"好", 2, ""}}),
		Pos{1, 2}},
}

func TestBufferCursor(t *testing.T) {
	for _, test := range bufferCursorTests {
		if p := test.buf.cursor(); p != test.want {
			t.Errorf("(%v).cursor() = %v, want %v", test.buf, p, test.want)
		}
	}
}

var buffersHeighTests = []struct {
	buffers []*buffer
	want    int
}{
	{nil, 0},
	{[]*buffer{newBuffer(10)}, 1},
	{[]*buffer{
		newBuffer(10).setLines([]cell{}, []cell{}),
		newBuffer(10),
		newBuffer(10).setLines([]cell{}, []cell{})},
		5},
}

func TestBuffersHeight(t *testing.T) {
	for _, test := range buffersHeighTests {
		if h := buffersHeight(test.buffers...); h != test.want {
			t.Errorf("buffersHeight(%v...) = %v, want %v",
				test.buffers, h, test.want)
		}
	}
}

var bufferWritesTests = []struct {
	buf   *buffer
	text  string
	style string
	want  *buffer
}{
	// Writing nothing.
	{newBuffer(10), "", "", newBuffer(10)},
	// Writing a single rune.
	{newBuffer(10), "a", "1", newBuffer(10).setLines([]cell{{"a", 1, "1"}})},
	// Writing control character.
	{newBuffer(10), "\033", "",
		newBuffer(10).setLines(
			[]cell{{"^[", 2, styleForControlChar.String()}},
		)},
	// Writing styled control character.
	{newBuffer(10), "a\033b", "1",
		newBuffer(10).setLines(
			[]cell{
				{"a", 1, "1"},
				{"^[", 2, "1;" + styleForControlChar.String()},
				{"b", 1, "1"},
			},
		)},
	// Writing text containing a newline.
	{newBuffer(10), "a\nb", "1",
		newBuffer(10).setLines(
			[]cell{{"a", 1, "1"}}, []cell{{"b", 1, "1"}},
		)},
	// Writing text containing a newline when there is indent.
	{newBuffer(10).setIndent(2), "a\nb", "1",
		newBuffer(10).setIndent(2).setLines(
			[]cell{{"a", 1, "1"}},
			[]cell{{" ", 1, ""}, {" ", 1, ""}, {"b", 1, "1"}},
		)},
	// Writing long text that triggers wrapping.
	{newBuffer(4), "aaaab", "1",
		newBuffer(4).setLines(
			[]cell{{"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}},
			[]cell{{"b", 1, "1"}},
		)},
	// Writing long text that triggers wrapping when there is indent.
	{newBuffer(4).setIndent(2), "aaaab", "1",
		newBuffer(4).setIndent(2).setLines(
			[]cell{{"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}},
			[]cell{{" ", 1, ""}, {" ", 1, ""}, {"b", 1, "1"}},
		)},
	// Writing long text that triggers eager wrapping.
	{newBuffer(4).setIndent(2).setEagerWrap(true), "aaaa", "1",
		newBuffer(4).setIndent(2).setEagerWrap(true).setLines(
			[]cell{{"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}},
			[]cell{{" ", 1, ""}, {" ", 1, ""}},
		)},
}

// TestBufferWrites tests buffer.writes by calling writes on a buffer and see if
// the buffer matches what is expected.
func TestBufferWrites(t *testing.T) {
	for _, test := range bufferWritesTests {
		b := test.buf
		b.writes(test.text, test.style)
		if !reflect.DeepEqual(b, test.want) {
			t.Errorf("buf.writes(%q, %q) makes it %v, want %v",
				test.text, test.style, b, test.want)
		}
	}
}

var bufferTrimToLinesTests = []struct {
	buf  *buffer
	low  int
	high int
	want *buffer
}{
	{
		newBuffer(10).setLines(
			[]cell{{"a", 1, ""}}, []cell{{"b", 1, ""}},
			[]cell{{"c", 1, ""}}, []cell{{"d", 1, ""}},
		), 0, 2,
		newBuffer(10).setLines(
			[]cell{{"a", 1, ""}}, []cell{{"b", 1, ""}},
		),
	},
	// With dot.
	{
		newBuffer(10).setLines(
			[]cell{{"a", 1, ""}}, []cell{{"b", 1, ""}},
			[]cell{{"c", 1, ""}}, []cell{{"d", 1, ""}},
		).setDot(Pos{1, 1}), 1, 3,
		newBuffer(10).setLines(
			[]cell{{"b", 1, ""}}, []cell{{"c", 1, ""}},
		).setDot(Pos{0, 1}),
	},
	// With dot that is going to be trimmed away.
	{
		newBuffer(10).setLines(
			[]cell{{"a", 1, ""}}, []cell{{"b", 1, ""}},
			[]cell{{"c", 1, ""}}, []cell{{"d", 1, ""}},
		).setDot(Pos{0, 1}), 1, 3,
		newBuffer(10).setLines(
			[]cell{{"b", 1, ""}}, []cell{{"c", 1, ""}},
		).setDot(Pos{0, 1}),
	},
}

func TestBufferTrimToLines(t *testing.T) {
	for _, test := range bufferTrimToLinesTests {
		b := test.buf
		b.trimToLines(test.low, test.high)
		if !reflect.DeepEqual(b, test.want) {
			t.Errorf("buf.trimToLines(%v, %v) makes it %v, want %v",
				test.low, test.high, b, test.want)
		}
	}
}

var bufferExtendTests = []struct {
	buf     *buffer
	buf2    *buffer
	moveDot bool
	want    *buffer
}{
	{
		newBuffer(10).setLines([]cell{{"a", 1, ""}}, []cell{{"b", 1, ""}}),
		newBuffer(11).setLines([]cell{{"c", 1, ""}}, []cell{{"d", 1, ""}}),
		false,
		newBuffer(10).setLines(
			[]cell{{"a", 1, ""}}, []cell{{"b", 1, ""}},
			[]cell{{"c", 1, ""}}, []cell{{"d", 1, ""}},
		),
	},
	// Moving dot.
	{
		newBuffer(10).setLines([]cell{{"a", 1, ""}}, []cell{{"b", 1, ""}}),
		newBuffer(11).setLines(
			[]cell{{"c", 1, ""}}, []cell{{"d", 1, ""}},
		).setDot(Pos{1, 1}),
		true,
		newBuffer(10).setLines(
			[]cell{{"a", 1, ""}}, []cell{{"b", 1, ""}},
			[]cell{{"c", 1, ""}}, []cell{{"d", 1, ""}},
		).setDot(Pos{3, 1}),
	},
}

func TestExtend(t *testing.T) {
	for _, test := range bufferExtendTests {
		b := test.buf
		b.extend(test.buf2, test.moveDot)
		if !reflect.DeepEqual(b, test.want) {
			t.Errorf("buf.extend(%v, %v) makes it %v, want %v",
				test.buf2, test.moveDot, b, test.want)
		}
	}
}

var bufferExtendRightTests = []struct {
	buf  *buffer
	buf2 *buffer
	w    int
	want *buffer
}{
	// No padding, equal height.
	{
		newBuffer(10).setLines([]cell{{"a", 1, ""}}, []cell{}),
		newBuffer(11).setLines([]cell{{"c", 1, ""}}, []cell{{"d", 1, ""}}),
		0,
		newBuffer(10).setLines(
			[]cell{{"a", 1, ""}, {"c", 1, ""}},
			[]cell{{"d", 1, ""}},
		),
	},
	// With padding.
	{
		newBuffer(10).setLines([]cell{{"a", 1, ""}}, []cell{{"b", 1, ""}}),
		newBuffer(11).setLines([]cell{{"c", 1, ""}}, []cell{{"d", 1, ""}}),
		2,
		newBuffer(10).setLines(
			[]cell{{"a", 1, ""}, {" ", 1, ""}, {"c", 1, ""}},
			[]cell{{"b", 1, ""}, {" ", 1, ""}, {"d", 1, ""}},
		),
	},
	// buf is higher.
	{
		newBuffer(10).setLines(
			[]cell{{"a", 1, ""}},
			[]cell{{"b", 1, ""}},
			[]cell{{"x", 1, ""}},
		),
		newBuffer(11).setLines([]cell{{"c", 1, ""}}, []cell{{"d", 1, ""}}),
		1,
		newBuffer(10).setLines(
			[]cell{{"a", 1, ""}, {"c", 1, ""}},
			[]cell{{"b", 1, ""}, {"d", 1, ""}},
			[]cell{{"x", 1, ""}},
		),
	},
	// buf2 is higher.
	{
		newBuffer(10).setLines([]cell{{"a", 1, ""}}, []cell{{"b", 1, ""}}),
		newBuffer(11).setLines(
			[]cell{{"c", 1, ""}}, []cell{{"d", 1, ""}}, []cell{{"e", 1, ""}},
		),
		1,
		newBuffer(10).setLines(
			[]cell{{"a", 1, ""}, {"c", 1, ""}},
			[]cell{{"b", 1, ""}, {"d", 1, ""}},
			[]cell{{" ", 1, ""}, {"e", 1, ""}},
		),
	},
}

func TestExtendRight(t *testing.T) {
	for _, test := range bufferExtendRightTests {
		b := test.buf
		b.extendRight(test.buf2, test.w)
		if !reflect.DeepEqual(b, test.want) {
			t.Errorf("buf.extendRight(%v, %v) makes it %v, want %v",
				test.buf2, test.w, b, test.want)
		}
	}
}
