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
		&Buffer{Width: 10, Lines: [][]Cell{{cell("a", "1")}}}},
	// Writing control character.
	{NewBufferBuilder(10), "\033", "",
		&Buffer{Width: 10, Lines: [][]Cell{{cell("^[", "7")}}}},
	// Writing styled control character.
	{NewBufferBuilder(10), "a\033b", "1",
		&Buffer{Width: 10, Lines: [][]Cell{{
			cell("a", "1"),
			cell("^[", "1;7"),
			cell("b", "1")}}}},
	// Writing text containing a newline.
	{NewBufferBuilder(10), "a\nb", "1",
		&Buffer{Width: 10, Lines: [][]Cell{
			{cell("a", "1")}, {cell("b", "1")}}}},
	// Writing text containing a newline when there is indent.
	{NewBufferBuilder(10).SetIndent(2), "a\nb", "1",
		&Buffer{Width: 10, Lines: [][]Cell{
			{cell("a", "1")},
			{cell(" ", ""), cell(" ", ""), cell("b", "1")},
		}}},
	// Writing long text that triggers wrapping.
	{NewBufferBuilder(4), "aaaab", "1",
		&Buffer{Width: 4, Lines: [][]Cell{
			{cell("a", "1"), cell("a", "1"), cell("a", "1"), cell("a", "1")},
			{cell("b", "1")}}}},
	// Writing long text that triggers wrapping when there is indent.
	{NewBufferBuilder(4).SetIndent(2), "aaaab", "1",
		&Buffer{Width: 4, Lines: [][]Cell{
			{cell("a", "1"), cell("a", "1"), cell("a", "1"), cell("a", "1")},
			{cell(" ", ""), cell(" ", ""), cell("b", "1")}}}},
	// Writing long text that triggers eager wrapping.
	{NewBufferBuilder(4).SetIndent(2).SetEagerWrap(true), "aaaa", "1",
		&Buffer{Width: 4, Lines: [][]Cell{
			{cell("a", "1"), cell("a", "1"), cell("a", "1"), cell("a", "1")},
			{cell(" ", ""), cell(" ", "")}}}},
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
			{cell("f", "4"), cell("o", "4"), cell("o", ""), cell(" ", "")},
			{cell("b", ""), cell("a", ""), cell("r", "")},
		}},
	},
}

func cloneBufferBuilder(bb *BufferBuilder) *BufferBuilder {
	return &BufferBuilder{
		Width:      bb.Width,
		Col:        bb.Col,
		Indent:     bb.Indent,
		EagerWrap:  bb.EagerWrap,
		Lines:      cloneLines(bb.Lines),
		Dot:        bb.Dot,
		PreIndent:  bb.PreIndent,
		PostIndent: bb.PostIndent,
	}
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
			cell("p", ""), cell("r", ""), cell("o", ""),
			cell("m", ""), cell("p", ""), cell("t", ""),
			{"\033]133;A\007", "", 0}, // zero-width
			cell(">", ""), cell(" ", ""),
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

func TestIndentHooks(t *testing.T) {
	tests := []struct {
		name      string
		indent    int
		text      string
		wantCalls []string
	}{
		{
			name:      "multiline with indent",
			indent:    2,
			text:      "line1\nline2\nline3",
			wantCalls: []string{"pre", "post", "pre", "post"},
		},
		{
			name:      "single line with indent",
			indent:    2,
			text:      "single line",
			wantCalls: nil,
		},
		{
			name:      "multiline without indent",
			indent:    0,
			text:      "line1\nline2",
			wantCalls: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bb := NewBufferBuilder(20).SetIndent(test.indent)

			var calls []string
			bb.PreIndent = func() { calls = append(calls, "pre") }
			bb.PostIndent = func() { calls = append(calls, "post") }

			bb.WriteStringSGR(test.text, "")

			if !reflect.DeepEqual(calls, test.wantCalls) {
				t.Errorf("hook calls = %v, want %v", calls, test.wantCalls)
			}
		})
	}
}
