package cliedit

import (
	"strings"
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/styled"
)

func TestPrompt_ValueOutput(t *testing.T) {
	ttyCtrl, cleanup := setupWithRC(
		`edit:prompt = { put 'val'; styled '> ' red }`)
	defer cleanup()

	ttyCtrl.TestBuffer(t,
		bb().WritePlain("val").WriteStyled(styled.MakeText("> ", "red")).
			SetDotToCursor().Buffer())
}

func TestPrompt_ByteOutput(t *testing.T) {
	ttyCtrl, cleanup := setupWithRC(`edit:prompt = { put 'bytes> ' }`)
	defer cleanup()

	ttyCtrl.TestBuffer(t,
		bb().WritePlain("bytes> ").SetDotToCursor().Buffer())
}

func TestPrompt_NotifiesInvalidValueOutput(t *testing.T) {
	ttyCtrl, cleanup := setupWithRC(`edit:prompt = { put good [bad] good2 }`)
	defer cleanup()

	ttyCtrl.TestBuffer(t,
		bb().WritePlain("goodgood2").SetDotToCursor().Buffer())
	ttyCtrl.TestNotesBuffer(t, bb().
		WritePlain("invalid output type from prompt: list").Buffer())
}

func TestPrompt_NotifiesException(t *testing.T) {
	ttyCtrl, cleanup := setupWithRC(`edit:prompt = { fail ERROR }`)
	defer cleanup()

	ttyCtrl.TestNotesBuffer(t, bb().
		WritePlain("prompt function error: ERROR").Buffer())
}

func TestRPrompt(t *testing.T) {
	ttyCtrl, cleanup := setupWithRC(`edit:rprompt = { put 'RRR' }`)
	defer cleanup()

	ttyCtrl.TestBuffer(t,
		bb().WritePlain("~> ").SetDotToCursor().
			WritePlain(strings.Repeat(" ", testTTYWidth-6)+"RRR").Buffer())
}

func TestPromptEagerness(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setup()
	defer cleanup()

	evalf(ev, `i = 0`)
	evalf(ev, `edit:prompt = { i = (+ $i 1); put $i'> ' }`)
	evalf(ev, `edit:-prompt-eagerness = 10`)
	_, _, stop := start(ed)
	defer stop()

	wantBuf1 := bb().WritePlain("1> ").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf1)
	// With eagerness = 10, any key press will cause the prompt to be
	// recomputed.
	ttyCtrl.Inject(term.K(ui.Backspace))
	wantBuf2 := bb().WritePlain("2> ").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf2)
}

func TestPromptStaleThreshold(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setup()
	defer cleanup()

	evalf(ev, `edit:prompt = { esleep 0.1; put '> ' }`)
	evalf(ev, `edit:prompt-stale-threshold = 0.05`)
	_, _, stop := start(ed)
	defer stop()

	wantBufStale := bb().
		WriteStyled(styled.MakeText("???> ", "inverse")).SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBufStale)

	wantBufFresh := bb().WritePlain("> ").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBufFresh)
}

func TestPromptStaleTransform(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setup()
	defer cleanup()

	evalf(ev, `edit:prompt = { esleep 0.1; put '> ' }`)
	evalf(ev, `edit:prompt-stale-threshold = 0.05`)
	evalf(ev, `edit:prompt-stale-transform = [a]{ put S; put $a; put S }`)
	_, _, stop := start(ed)
	defer stop()

	wantBufStale := bb().
		WriteStyled(styled.Plain("S???> S")).SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBufStale)
}

func TestDefaultPromptForNonRoot(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setup()
	defer cleanup()
	ev.Global["f"] = vars.NewReadOnly(getDefaultPrompt(false))
	evalf(ev, `edit:prompt = $f`)

	_, _, stop := start(ed)
	defer stop()

	wantBuf := bb().WritePlain("~> ").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
}

func TestDefaultPromptForRoot(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setup()
	defer cleanup()
	ev.Global["f"] = vars.NewReadOnly(getDefaultPrompt(true))
	evalf(ev, `edit:prompt = $f`)

	_, _, stop := start(ed)
	defer stop()

	wantBuf := bb().WritePlain("~").
		WriteStyled(styled.MakeText("# ", "red")).SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
}

func TestDefaultRPrompt(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setup()
	defer cleanup()
	ev.Global["f"] = vars.NewReadOnly(getDefaultRPrompt("elf", "host"))
	evalf(ev, `edit:rprompt = $f`)

	_, _, stop := start(ed)
	defer stop()

	wantBuf := bb().WritePlain("~> ").SetDotToCursor().
		WritePlain(strings.Repeat(" ", 49)).
		WriteStyled(styled.MakeText("elf@host", "inverse")).Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
}

func setupWithRC(codes ...string) (cli.TTYCtrl, func()) {
	ed, ttyCtrl, ev, cleanup := setup()
	for _, code := range codes {
		evalf(ev, `%s`, code)
	}
	_, _, stop := start(ed)
	return ttyCtrl, func() {
		stop()
		cleanup()
	}
}
