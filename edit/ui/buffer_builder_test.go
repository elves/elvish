package ui

import (
	"reflect"
	"testing"
)

var bufferBuilderWritesTests = []struct {
	bb    *BufferBuilder
	text  string
	style string
	want  *Buffer
}{
	// Writing nothing.
	{NewBufferBuilder(10), "", "", &Buffer{Width: 10, Lines: Lines{Line{}}}},
	// Writing a single rune.
	{NewBufferBuilder(10), "a", "1",
		&Buffer{Width: 10, Lines: Lines{Line{C("a", "1")}}}},
	// Writing control character.
	{NewBufferBuilder(10), "\033", "",
		&Buffer{Width: 10, Lines: Lines{
			Line{C("^[", styleForControlChar.String())}}}},
	// Writing styled control character.
	{NewBufferBuilder(10), "a\033b", "1",
		&Buffer{Width: 10, Lines: Lines{Line{
			C("a", "1"),
			C("^[", "1;"+styleForControlChar.String()),
			C("b", "1")}}}},
	// Writing text containing a newline.
	{NewBufferBuilder(10), "a\nb", "1",
		&Buffer{Width: 10, Lines: Lines{
			Line{C("a", "1")}, Line{C("b", "1")}}}},
	// Writing text containing a newline when there is indent.
	{NewBufferBuilder(10).SetIndent(2), "a\nb", "1",
		&Buffer{Width: 10, Lines: Lines{
			Line{C("a", "1")},
			Line{C(" ", ""), C(" ", ""), C("b", "1")},
		}}},
	// Writing long text that triggers wrapping.
	{NewBufferBuilder(4), "aaaab", "1",
		&Buffer{Width: 4, Lines: Lines{
			Line{C("a", "1"), C("a", "1"), C("a", "1"), C("a", "1")},
			Line{C("b", "1")}}}},
	// Writing long text that triggers wrapping when there is indent.
	{NewBufferBuilder(4).SetIndent(2), "aaaab", "1",
		&Buffer{Width: 4, Lines: Lines{
			Line{C("a", "1"), C("a", "1"), C("a", "1"), C("a", "1")},
			Line{C(" ", ""), C(" ", ""), C("b", "1")}}}},
	// Writing long text that triggers eager wrapping.
	{NewBufferBuilder(4).SetIndent(2).SetEagerWrap(true), "aaaa", "1",
		&Buffer{Width: 4, Lines: Lines{
			Line{C("a", "1"), C("a", "1"), C("a", "1"), C("a", "1")},
			Line{C(" ", ""), C(" ", "")}}}},
}

// TestBufferWrites tests BufferBuilder.Writes by calling Writes on a
// BufferBuilder and see if the built Buffer matches what is expected.
func TestBufferBuilderWrites(t *testing.T) {
	for _, test := range bufferBuilderWritesTests {
		bb := test.bb
		bb.WriteStringSGR(test.text, test.style)
		buf := bb.Buffer()
		if !reflect.DeepEqual(buf, test.want) {
			t.Errorf("buf.writes(%q, %q) makes it %v, want %v",
				test.text, test.style, buf, test.want)
		}
	}
}

var bufferBuilderExtendTests = []struct {
	bb      *BufferBuilder
	buf2    *Buffer
	moveDot bool
	want    *Buffer
}{
	{
		NewBufferBuilder(10).SetLines(Line{C("a", "")}, Line{C("b", "")}),
		&Buffer{Width: 11, Lines: Lines{Line{C("c", "")}, Line{C("d", "")}}},
		false,
		&Buffer{Width: 10, Lines: Lines{
			Line{C("a", "")}, Line{C("b", "")},
			Line{C("c", "")}, Line{C("d", "")}}},
	},
	// Moving dot.
	{
		NewBufferBuilder(10).SetLines(Line{C("a", "")}, Line{C("b", "")}),
		&Buffer{
			Width: 11,
			Lines: Lines{Line{C("c", "")}, Line{C("d", "")}},
			Dot:   Pos{1, 1},
		},
		true,
		&Buffer{
			Width: 10,
			Lines: Lines{
				Line{C("a", "")}, Line{C("b", "")},
				Line{C("c", "")}, Line{C("d", "")}},
			Dot: Pos{3, 1},
		},
	},
}

func TestBufferBuilderExtend(t *testing.T) {
	for _, test := range bufferBuilderExtendTests {
		bb := test.bb
		bb.Extend(test.buf2, test.moveDot)
		b := bb.Buffer()
		if !reflect.DeepEqual(b, test.want) {
			t.Errorf("buf.extend(%v, %v) makes it %v, want %v",
				test.buf2, test.moveDot, b, test.want)
		}
	}
}

var bufferBuilderExtendRightTests = []struct {
	bb   *BufferBuilder
	buf2 *Buffer
	w    int
	want *Buffer
}{
	// No padding, equal height.
	{
		NewBufferBuilder(10).SetLines(Line{C("a", "")}, Line{}),
		&Buffer{Width: 11, Lines: Lines{Line{C("c", "")}, Line{C("d", "")}}},
		0,
		&Buffer{Width: 10, Lines: Lines{
			Line{C("a", ""), C("c", "")}, Line{C("d", "")}}},
	},
	// With padding.
	{
		NewBufferBuilder(10).SetLines(Line{C("a", "")}, Line{C("b", "")}),
		&Buffer{Width: 11, Lines: Lines{Line{C("c", "")}, Line{C("d", "")}}},
		2,
		&Buffer{Width: 10, Lines: Lines{
			Line{C("a", ""), C(" ", ""), C("c", "")},
			Line{C("b", ""), C(" ", ""), C("d", "")}}},
	},
	// buf is higher.
	{
		NewBufferBuilder(10).SetLines(
			Line{C("a", "")}, Line{C("b", "")}, Line{C("x", "")}),
		&Buffer{Width: 11, Lines: Lines{
			Line{C("c", "")}, Line{C("d", "")},
		}},
		1,
		&Buffer{Width: 10, Lines: Lines{
			Line{C("a", ""), C("c", "")},
			Line{C("b", ""), C("d", "")},
			Line{C("x", "")}}},
	},
	// buf2 is higher.
	{
		NewBufferBuilder(10).SetLines(
			Line{C("a", "")}, Line{C("b", "")}),
		&Buffer{Width: 11, Lines: Lines{
			Line{C("c", "")}, Line{C("d", "")}, Line{C("e", "")},
		}},
		1,
		&Buffer{Width: 10, Lines: Lines{
			Line{C("a", ""), C("c", "")},
			Line{C("b", ""), C("d", "")},
			Line{C(" ", ""), C("e", "")}}},
	},
}

func TestBufferBuilderExtendRight(t *testing.T) {
	for _, test := range bufferBuilderExtendRightTests {
		bb := test.bb
		bb.ExtendRight(test.buf2, test.w)
		b := bb.Buffer()
		if !reflect.DeepEqual(b, test.want) {
			t.Errorf("buf.extendRight(%v, %v) makes it %v, want %v",
				test.buf2, test.w, b, test.want)
		}
	}
}
