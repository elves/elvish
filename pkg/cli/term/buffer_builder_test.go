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
		&Buffer{Width: 10, Lines: [][]Cell{{{"a", "1", false}}}}},
	// Writing control character.
	{NewBufferBuilder(10), "\033", "",
		&Buffer{Width: 10, Lines: [][]Cell{{{"^[", "7", false}}}}},
	// Writing styled control character.
	{NewBufferBuilder(10), "a\033b", "1",
		&Buffer{Width: 10, Lines: [][]Cell{{
			{"a", "1", false},
			{"^[", "1;7", false},
			{"b", "1", false}}}}},
	// Writing text containing a newline.
	{NewBufferBuilder(10), "a\nb", "1",
		&Buffer{Width: 10, Lines: [][]Cell{
			{{"a", "1", false}}, {{"b", "1", false}}}}},
	// Writing text containing a newline when there is indent.
	{NewBufferBuilder(10).SetIndent(2), "a\nb", "1",
		&Buffer{Width: 10, Lines: [][]Cell{
			{{"a", "1", false}},
			{{" ", "", false}, {" ", "", false}, {"b", "1", false}},
		}}},
	// Writing long text that triggers wrapping.
	{NewBufferBuilder(4), "aaaab", "1",
		&Buffer{Width: 4, Lines: [][]Cell{
			{{"a", "1", false}, {"a", "1", false}, {"a", "1", false}, {"a", "1", false}},
			{{"b", "1", false}}}}},
	// Writing long text that triggers wrapping when there is indent.
	{NewBufferBuilder(4).SetIndent(2), "aaaab", "1",
		&Buffer{Width: 4, Lines: [][]Cell{
			{{"a", "1", false}, {"a", "1", false}, {"a", "1", false}, {"a", "1", false}},
			{{" ", "", false}, {" ", "", false}, {"b", "1", false}}}}},
	// Writing long text that triggers eager wrapping.
	{NewBufferBuilder(4).SetIndent(2).SetEagerWrap(true), "aaaa", "1",
		&Buffer{Width: 4, Lines: [][]Cell{
			{{"a", "1", false}, {"a", "1", false}, {"a", "1", false}, {"a", "1", false}},
			{{" ", "", false}, {" ", "", false}}}}},
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
			{{"f", "4", false}, {"o", "4", false}, {"o", "", false}, {" ", "", false}},
			{{"b", "", false}, {"a", "", false}, {"r", "", false}},
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

func TestWriteZeroWidth(t *testing.T) {
	bb := NewBufferBuilder(10)
	bb.Write("prompt").WriteZeroWidth("\033]133;A\007").Write("> ")
	buf := bb.Buffer()

	want := &Buffer{
		Width: 10,
		Lines: [][]Cell{{
			{"p", "", false}, {"r", "", false}, {"o", "", false},
			{"m", "", false}, {"p", "", false}, {"t", "", false},
			{"\033]133;A\007", "", true}, // zero-width
			{">", "", false}, {" ", "", false},
		}},
	}

	if !reflect.DeepEqual(buf, want) {
		t.Errorf("Got buf:\n%v\nWant:\n%v", buf, want)
	}

	// Verify width calculation excludes zero-width cells
	if width := cellsWidth(buf.Lines[0]); width != 8 {
		t.Errorf("cellsWidth = %d, want 8", width)
	}

	// Verify cursor position (zero-width doesn't affect it)
	if bb.Col != 8 {
		t.Errorf("bb.Col = %d, want 8", bb.Col)
	}
}

func cloneBufferBuilder(bb *BufferBuilder) *BufferBuilder {
	return &BufferBuilder{
		bb.Width, bb.Col, bb.Indent,
		bb.EagerWrap, cloneLines(bb.Lines), bb.Dot}
}
