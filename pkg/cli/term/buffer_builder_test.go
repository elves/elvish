package term

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/ui"
)

var bufferBuilderWritesTests = []struct {
	bb    *BufferBuilder
	text  string
	style string
	want  *Buffer
}{
	// Writing nothing.
	{NewBufferBuilder(10), "", "", &Buffer{Width: 10, Lines: [][]Cell{{}}}},
	// Writing a single rune.
	{NewBufferBuilder(10), "a", "1",
		&Buffer{Width: 10, Lines: [][]Cell{{{"a", "1"}}}}},
	// Writing control character.
	{NewBufferBuilder(10), "\033", "",
		&Buffer{Width: 10, Lines: [][]Cell{{{"^[", "7"}}}}},
	// Writing styled control character.
	{NewBufferBuilder(10), "a\033b", "1",
		&Buffer{Width: 10, Lines: [][]Cell{{
			{"a", "1"},
			{"^[", "1;7"},
			{"b", "1"}}}}},
	// Writing text containing a newline.
	{NewBufferBuilder(10), "a\nb", "1",
		&Buffer{Width: 10, Lines: [][]Cell{
			{{"a", "1"}}, {{"b", "1"}}}}},
	// Writing text containing a newline when there is indent.
	{NewBufferBuilder(10).SetIndent(2), "a\nb", "1",
		&Buffer{Width: 10, Lines: [][]Cell{
			{{"a", "1"}},
			{{" ", ""}, {" ", ""}, {"b", "1"}},
		}}},
	// Writing long text that triggers wrapping.
	{NewBufferBuilder(4), "aaaab", "1",
		&Buffer{Width: 4, Lines: [][]Cell{
			{{"a", "1"}, {"a", "1"}, {"a", "1"}, {"a", "1"}},
			{{"b", "1"}}}}},
	// Writing long text that triggers wrapping when there is indent.
	{NewBufferBuilder(4).SetIndent(2), "aaaab", "1",
		&Buffer{Width: 4, Lines: [][]Cell{
			{{"a", "1"}, {"a", "1"}, {"a", "1"}, {"a", "1"}},
			{{" ", ""}, {" ", ""}, {"b", "1"}}}}},
	// Writing long text that triggers eager wrapping.
	{NewBufferBuilder(4).SetIndent(2).SetEagerWrap(true), "aaaa", "1",
		&Buffer{Width: 4, Lines: [][]Cell{
			{{"a", "1"}, {"a", "1"}, {"a", "1"}, {"a", "1"}},
			{{" ", ""}, {" ", ""}}}}},
}

// TestBufferBuilderWrites tests BufferBuilder.Writes by calling Writes on a
// BufferBuilder and see if the built Buffer matches what is expected.
func TestBufferBuilderWrites(t *testing.T) {
	for _, test := range bufferBuilderWritesTests {
		bb := cloneBufferBuilder(test.bb)
		bb.WriteStringSGR(test.text, test.style)
		buf := bb.Buffer()
		if !reflect.DeepEqual(buf, test.want) {
			t.Errorf("buf.writes(%q, %q) makes it %v, want %v",
				test.text, test.style, buf, test.want)
		}
	}
}

var styles = ui.RuneStylesheet{
	'-': ui.Underlined,
}

var bufferBuilderTests = []struct {
	name    string
	builder *BufferBuilder
	wantBuf *Buffer
}{
	{
		"MarkLines",
		NewBufferBuilder(10).MarkLines(
			"foo ", styles,
			"--  ", DotHere, "\n",
			"",
			"bar",
		),
		&Buffer{Width: 10, Dot: Pos{0, 4}, Lines: [][]Cell{
			{{"f", "4"}, {"o", "4"}, {"o", ""}, {" ", ""}},
			{{"b", ""}, {"a", ""}, {"r", ""}},
		}},
	},
}

func TestBufferBuilder(t *testing.T) {
	for _, test := range bufferBuilderTests {
		t.Run(test.name, func(t *testing.T) {
			buf := test.builder.Buffer()
			if !reflect.DeepEqual(buf, test.wantBuf) {
				t.Errorf("Got buf %v, want %v", buf, test.wantBuf)
			}
		})
	}
}

func cloneBufferBuilder(bb *BufferBuilder) *BufferBuilder {
	return &BufferBuilder{
		bb.Width, bb.Col, bb.Indent,
		bb.EagerWrap, cloneLines(bb.Lines), bb.Dot}
}
