package main

import (
	"unicode/utf8"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/strutil"
	"src.elv.sh/pkg/ui"
)

var TextArea = etk.WithInit(comps.TextArea,
	"prompt", ui.T("~> "),
	"abbr", func(y func(a, f string)) { y("foo", "lorem") },
	"binding",
	func(ev term.Event, c etk.Context, tag string, f etk.React) etk.Reaction {
		reaction := f(ev)
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
	})

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
