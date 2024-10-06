package comps_test

import (
	"embed"
	"testing"
	"unicode/utf8"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/edit/highlight"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/etk/etktest"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/strutil"
	"src.elv.sh/pkg/ui"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	hl := highlight.NewHighlighter(highlight.Config{
		HasCommand: func(cmd string) bool { return cmd == "echo" },
	})

	evaltest.TestTranscriptsInFS(t, transcripts,
		"text-area-fixture", etktest.MakeFixture(comps.TextArea),
		"text-area-demo-fixture", etktest.MakeFixture(
			etk.WithInit(comps.TextArea,
				"binding", func(ev term.Event, c etk.Context, r etk.React) etk.Reaction {
					reaction := r(ev)
					if reaction != etk.Unused {
						return reaction
					}
					bufferVar := etk.BindState(c, "buffer", comps.TextBuffer{})
					switch ev {
					case term.K(ui.Left):
						bufferVar.Swap(makeMove(moveDotLeft))
					case term.K(ui.Right):
						bufferVar.Swap(makeMove(moveDotRight))
					case term.K(ui.Home):
						bufferVar.Swap(makeMove(moveDotSOL))
					case term.K(ui.End):
						bufferVar.Swap(makeMove(moveDotEOL))
					default:
						return etk.Unused
					}
					return etk.Consumed
				},
				"highlighter", hl.Get,
			)),
		"abbr-table-in-global", evaltest.GoFnInGlobal("abbr-table",
			func(m vals.Map) func(f func(a, f string)) {
				return func(f func(a, f string)) {
					for it := m.Iterator(); it.HasElem(); it.Next() {
						k, v := it.Elem()
						f(vals.ToString(k), vals.ToString(v))
					}
				}
			}),
	)
}

// For demo

func makeMove(m func(string, int) int) func(comps.TextBuffer) comps.TextBuffer {
	return func(buf comps.TextBuffer) comps.TextBuffer {
		buf.Dot = m(buf.Content, buf.Dot)
		return buf
	}
}

func moveDotLeft(buffer string, dot int) int {
	_, w := utf8.DecodeLastRuneInString(buffer[:dot])
	return dot - w
}

func moveDotRight(buffer string, dot int) int {
	_, w := utf8.DecodeRuneInString(buffer[dot:])
	return dot + w
}

func moveDotSOL(buffer string, dot int) int {
	return strutil.FindLastSOL(buffer[:dot])
}

func moveDotEOL(buffer string, dot int) int {
	return strutil.FindFirstEOL(buffer[dot:]) + dot
}
