package term

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/wcwidth"
)

func cell(text, style string) Cell {
	return Cell{Text: text, Style: style, Width: wcwidth.Of(text)}
}

var cellsWidthTests = []struct {
	cells     []Cell
	wantWidth int
}{
	{[]Cell{}, 0},
	{[]Cell{cell("a", ""), cell("好", "")}, 3},
	{[]Cell{{"\033]133;A\007", "", 0}}, 0},
	{[]Cell{cell("a", ""), {"\033]133;B\007", "", 0}, cell("b", "")}, 2},
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
	want []Cell
}{
	{0, []Cell{}},
	{1, []Cell{cell(" ", "")}},
	{4, []Cell{cell(" ", ""), cell(" ", ""), cell(" ", ""), cell(" ", "")}},
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
	{[]Cell{}, []Cell{cell("a", "")}, false, 0},
	{
		[]Cell{cell("a", ""), cell("好", ""), cell("b", "")},
		[]Cell{cell("a", ""), cell("好", ""), cell("c", "")},
		false, 2,
	},
	{
		[]Cell{cell("a", ""), cell("好", ""), cell("b", "")},
		[]Cell{cell("a", ""), cell("好", "1"), cell("c", "")},
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
	buf  *Buffer
	want Pos
}{
	{
		&Buffer{Width: 10, Lines: [][]Cell{{}}},
		Pos{0, 0},
	},
	{
		&Buffer{Width: 10, Lines: [][]Cell{{cell("a", "")}, {cell("好", "")}}},
		Pos{1, 2},
	},
}

func TestEndPos(t *testing.T) {
	for _, test := range bufferCursorTests {
		if p := endPos(test.buf); p != test.want {
			t.Errorf("(%v).cursor() = %v, want %v", test.buf, p, test.want)
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
		&Buffer{Width: 10, Lines: [][]Cell{
			{cell("a", "")}, {cell("b", "")}, {cell("c", "")}, {cell("d", "")},
		}},
		0, 2,
		&Buffer{Width: 10, Lines: [][]Cell{
			{cell("a", "")}, {cell("b", "")},
		}},
	},
	// Negative low is treated as 0.

	{
		&Buffer{Width: 10, Lines: [][]Cell{
			{cell("a", "")}, {cell("b", "")}, {cell("c", "")}, {cell("d", "")},
		}},
		-1, 2,
		&Buffer{Width: 10, Lines: [][]Cell{
			{cell("a", "")}, {cell("b", "")},
		}},
	},
	// With dot.
	{
		&Buffer{Width: 10, Lines: [][]Cell{
			{cell("a", "")}, {cell("b", "")}, {cell("c", "")}, {cell("d", "")},
		}, Dot: Pos{1, 1}},
		1, 3,
		&Buffer{Width: 10, Lines: [][]Cell{
			{cell("b", "")}, {cell("c", "")},
		}, Dot: Pos{0, 1}},
	},
	// With dot that is going to be trimmed away.
	{
		&Buffer{Width: 10, Lines: [][]Cell{
			{cell("a", "")}, {cell("b", "")}, {cell("c", "")}, {cell("d", "")},
		}, Dot: Pos{0, 1}},
		1, 3,
		&Buffer{Width: 10, Lines: [][]Cell{
			{cell("b", "")}, {cell("c", "")},
		}, Dot: Pos{0, 1}},
	},
}

func TestBufferTrimToLines(t *testing.T) {
	for _, test := range bufferTrimToLinesTests {
		b := cloneBuffer(test.buf)
		b.TrimToLines(test.low, test.high)
		if !reflect.DeepEqual(b, test.want) {
			t.Errorf("buf.trimToLines(%v, %v) makes it %v, want %v",
				test.low, test.high, b, test.want)
		}
	}
}

var bufferExtendDownTests = []struct {
	buf     *Buffer
	buf2    *Buffer
	moveDot bool
	want    *Buffer
}{
	{
		&Buffer{Width: 10, Lines: [][]Cell{
			{cell("a", "")}, {cell("b", "")}}},
		&Buffer{Width: 11, Lines: [][]Cell{
			{cell("c", "")}, {cell("d", "")}}},
		false,
		&Buffer{Width: 11, Lines: [][]Cell{
			{cell("a", "")}, {cell("b", "")},
			{cell("c", "")}, {cell("d", "")}}},
	},
	// Moving dot.
	{
		&Buffer{Width: 10, Lines: [][]Cell{
			{cell("a", "")}, {cell("b", "")}}},
		&Buffer{
			Width: 11,
			Lines: [][]Cell{{cell("c", "")}, {cell("d", "")}},
			Dot:   Pos{1, 1},
		},
		true,
		&Buffer{
			Width: 11,
			Lines: [][]Cell{
				{cell("a", "")}, {cell("b", "")},
				{cell("c", "")}, {cell("d", "")}},
			Dot: Pos{3, 1},
		},
	},
}

func TestBufferExtendDown(t *testing.T) {
	for _, test := range bufferExtendDownTests {
		buf := cloneBuffer(test.buf)
		buf.ExtendDown(test.buf2, test.moveDot)
		if !reflect.DeepEqual(buf, test.want) {
			t.Errorf("buf.ExtendDown(%v, %v) makes it %v, want %v",
				test.buf2, test.moveDot, buf, test.want)
		}
	}
}

var bufferExtendRightTests = []struct {
	buf     *Buffer
	buf2    *Buffer
	moveDot bool
	want    *Buffer
}{
	// No padding, equal height.
	{
		&Buffer{Width: 1, Lines: [][]Cell{{cell("a", "")}, {cell("b", "")}}},
		&Buffer{Width: 1, Lines: [][]Cell{{cell("c", "")}, {cell("d", "")}}},
		false,
		&Buffer{Width: 2, Lines: [][]Cell{
			{cell("a", ""), cell("c", "")},
			{cell("b", ""), cell("d", "")}}},
	},
	// With padding, equal height.
	{
		&Buffer{Width: 2, Lines: [][]Cell{{cell("a", "")}, {cell("b", "")}}},
		&Buffer{Width: 1, Lines: [][]Cell{{cell("c", "")}, {cell("d", "")}}},
		false,
		&Buffer{Width: 3, Lines: [][]Cell{
			{cell("a", ""), cell(" ", ""), cell("c", "")},
			{cell("b", ""), cell(" ", ""), cell("d", "")}}},
	},
	// buf is higher.
	{
		&Buffer{Width: 1, Lines: [][]Cell{
			{cell("a", "")}, {cell("b", "")}, {cell("x", "")}}},
		&Buffer{Width: 1, Lines: [][]Cell{
			{cell("c", "")}, {cell("d", "")},
		}},
		false,
		&Buffer{Width: 2, Lines: [][]Cell{
			{cell("a", ""), cell("c", "")},
			{cell("b", ""), cell("d", "")},
			{cell("x", "")}}},
	},
	// buf2 is higher.
	{
		&Buffer{Width: 1, Lines: [][]Cell{{cell("a", "")}, {cell("b", "")}}},
		&Buffer{Width: 1, Lines: [][]Cell{
			{cell("c", "")}, {cell("d", "")}, {cell("e", "")},
		}},
		false,
		&Buffer{Width: 2, Lines: [][]Cell{
			{cell("a", ""), cell("c", "")},
			{cell("b", ""), cell("d", "")},
			{cell(" ", ""), cell("e", "")}}},
	},
	// Moving the dot.
	{
		&Buffer{Width: 1, Lines: [][]Cell{{cell("a", "")}, {cell("b", "")}}},
		&Buffer{Width: 1, Lines: [][]Cell{{cell("c", "")}, {cell("d", "")}}},
		true,
		&Buffer{Width: 2, Dot: Pos{0, 1}, Lines: [][]Cell{
			{cell("a", ""), cell("c", "")},
			{cell("b", ""), cell("d", "")}}},
	},
}

func TestBufferExtendRight(t *testing.T) {
	for _, test := range bufferExtendRightTests {
		buf := cloneBuffer(test.buf)
		buf.ExtendRight(test.buf2, test.moveDot)
		if !reflect.DeepEqual(buf, test.want) {
			t.Errorf("buf.extendRight(%v) makes it %v, want %v",
				test.buf2, buf, test.want)
		}
	}
}

func TestBufferBuffer(t *testing.T) {
	b := NewBufferBuilder(4).Write("text").Buffer()
	if b.Buffer() != b {
		t.Errorf("Buffer did not return itself")
	}
}

var bufferTTYStringTests = []struct {
	buf  *Buffer
	want string
}{
	{
		nil,
		"nil",
	},
	{
		NewBufferBuilder(4).
			Write("ABCD").
			Newline().
			Write("XY").
			Buffer(),
		"Width = 4, Dot = (0, 0)\n" +
			"┌────┐\n" +
			"│ABCD│\n" +
			"│XY$ │\n" +
			"└────┘\n",
	},
	{
		NewBufferBuilder(4).
			Write("A").SetDotHere().
			WriteStringSGR("B", "1").
			WriteStringSGR("C", "7").
			Write("D").
			Newline().
			WriteStringSGR("XY", "7").
			Buffer(),
		"Width = 4, Dot = (0, 1)\n" +
			"┌────┐\n" +
			"│A\033[1mB\033[;7mC\033[mD│\n" +
			"│\033[7mXY\033[m$ │\n" +
			"└────┘\n",
	},
}

func TestBufferTTYString(t *testing.T) {
	for _, test := range bufferTTYStringTests {
		ttyString := test.buf.TTYString()
		if ttyString != test.want {
			t.Errorf("TTYString -> %q, want %q", ttyString, test.want)
		}
	}
}

func cloneBuffer(b *Buffer) *Buffer {
	return &Buffer{b.Width, cloneLines(b.Lines), b.Dot}
}

func cloneLines(lines [][]Cell) [][]Cell {
	newLines := make([][]Cell, len(lines))
	for i, line := range lines {
		if line != nil {
			newLines[i] = make([]Cell, len(line))
			copy(newLines[i], line)
		}
	}
	return newLines
}
