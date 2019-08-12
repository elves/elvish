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
	{NewBufferBuilder(10), "", "", NewBuffer(10)},
	// Writing a single rune.
	{NewBufferBuilder(10), "a", "1", NewBuffer(10).SetLines([]Cell{{"a", 1, "1"}})},
	// Writing control character.
	{NewBufferBuilder(10), "\033", "",
		NewBuffer(10).SetLines(
			[]Cell{{"^[", 2, styleForControlChar.String()}},
		)},
	// Writing styled control character.
	{NewBufferBuilder(10), "a\033b", "1",
		NewBuffer(10).SetLines(
			[]Cell{
				{"a", 1, "1"},
				{"^[", 2, "1;" + styleForControlChar.String()},
				{"b", 1, "1"},
			},
		)},
	// Writing text containing a newline.
	{NewBufferBuilder(10), "a\nb", "1",
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, "1"}}, []Cell{{"b", 1, "1"}},
		)},
	// Writing text containing a newline when there is indent.
	{NewBufferBuilder(10).SetIndent(2), "a\nb", "1",
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, "1"}},
			[]Cell{{" ", 1, ""}, {" ", 1, ""}, {"b", 1, "1"}},
		)},
	// Writing long text that triggers wrapping.
	{NewBufferBuilder(4), "aaaab", "1",
		NewBuffer(4).SetLines(
			[]Cell{{"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}},
			[]Cell{{"b", 1, "1"}},
		)},
	// Writing long text that triggers wrapping when there is indent.
	{NewBufferBuilder(4).SetIndent(2), "aaaab", "1",
		NewBuffer(4).SetLines(
			[]Cell{{"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}},
			[]Cell{{" ", 1, ""}, {" ", 1, ""}, {"b", 1, "1"}},
		)},
	// Writing long text that triggers eager wrapping.
	{NewBufferBuilder(4).SetIndent(2).SetEagerWrap(true), "aaaa", "1",
		NewBuffer(4).SetLines(
			[]Cell{{"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}, {"a", 1, "1"}},
			[]Cell{{" ", 1, ""}, {" ", 1, ""}},
		)},
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
		NewBufferBuilder(10).SetLines(
			[]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}}),
		NewBuffer(11).SetLines([]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}}),
		false,
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}},
			[]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}}),
	},
	// Moving dot.
	{
		NewBufferBuilder(10).SetLines(
			[]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}}),
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
		NewBufferBuilder(10).SetLines([]Cell{{"a", 1, ""}}, []Cell{}),
		NewBuffer(11).SetLines([]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}}),
		0,
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}, {"c", 1, ""}},
			[]Cell{{"d", 1, ""}},
		),
	},
	// With padding.
	{
		NewBufferBuilder(10).SetLines(
			[]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}}),
		NewBuffer(11).SetLines([]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}}),
		2,
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}, {" ", 1, ""}, {"c", 1, ""}},
			[]Cell{{"b", 1, ""}, {" ", 1, ""}, {"d", 1, ""}},
		),
	},
	// buf is higher.
	{
		NewBufferBuilder(10).SetLines(
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
		NewBufferBuilder(10).SetLines(
			[]Cell{{"a", 1, ""}}, []Cell{{"b", 1, ""}}),
		NewBuffer(11).SetLines(
			[]Cell{{"c", 1, ""}}, []Cell{{"d", 1, ""}}, []Cell{{"e", 1, ""}}),
		1,
		NewBuffer(10).SetLines(
			[]Cell{{"a", 1, ""}, {"c", 1, ""}},
			[]Cell{{"b", 1, ""}, {"d", 1, ""}},
			[]Cell{{" ", 1, ""}, {"e", 1, ""}},
		),
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
