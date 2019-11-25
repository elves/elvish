package term

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
		&Buffer{Width: 10, Lines: Lines{Line{Cell{"a", "1"}}}}},
	// Writing control character.
	{NewBufferBuilder(10), "\033", "",
		&Buffer{Width: 10, Lines: Lines{Line{Cell{"^[", "7"}}}}},
	// Writing styled control character.
	{NewBufferBuilder(10), "a\033b", "1",
		&Buffer{Width: 10, Lines: Lines{Line{
			Cell{"a", "1"},
			Cell{"^[", "1;7"},
			Cell{"b", "1"}}}}},
	// Writing text containing a newline.
	{NewBufferBuilder(10), "a\nb", "1",
		&Buffer{Width: 10, Lines: Lines{
			Line{Cell{"a", "1"}}, Line{Cell{"b", "1"}}}}},
	// Writing text containing a newline when there is indent.
	{NewBufferBuilder(10).SetIndent(2), "a\nb", "1",
		&Buffer{Width: 10, Lines: Lines{
			Line{Cell{"a", "1"}},
			Line{Cell{" ", ""}, Cell{" ", ""}, Cell{"b", "1"}},
		}}},
	// Writing long text that triggers wrapping.
	{NewBufferBuilder(4), "aaaab", "1",
		&Buffer{Width: 4, Lines: Lines{
			Line{Cell{"a", "1"}, Cell{"a", "1"}, Cell{"a", "1"}, Cell{"a", "1"}},
			Line{Cell{"b", "1"}}}}},
	// Writing long text that triggers wrapping when there is indent.
	{NewBufferBuilder(4).SetIndent(2), "aaaab", "1",
		&Buffer{Width: 4, Lines: Lines{
			Line{Cell{"a", "1"}, Cell{"a", "1"}, Cell{"a", "1"}, Cell{"a", "1"}},
			Line{Cell{" ", ""}, Cell{" ", ""}, Cell{"b", "1"}}}}},
	// Writing long text that triggers eager wrapping.
	{NewBufferBuilder(4).SetIndent(2).SetEagerWrap(true), "aaaa", "1",
		&Buffer{Width: 4, Lines: Lines{
			Line{Cell{"a", "1"}, Cell{"a", "1"}, Cell{"a", "1"}, Cell{"a", "1"}},
			Line{Cell{" ", ""}, Cell{" ", ""}}}}},
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
