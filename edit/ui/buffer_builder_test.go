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
		bb.WriteString(test.text, test.style)
		buf := bb.Buffer()
		if !reflect.DeepEqual(buf, test.want) {
			t.Errorf("buf.writes(%q, %q) makes it %v, want %v",
				test.text, test.style, buf, test.want)
		}
	}
}
