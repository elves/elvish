package cliedit

import (
	"strings"
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/styled"
)

func TestPrompt_ValueOutput(t *testing.T) {
	ttyCtrl, _, cleanup := setupWithRC(
		`edit:prompt = { put 'val'; styled '> ' red }`)
	defer cleanup()

	ttyCtrl.TestBuffer(t,
		bb().WritePlain("val").WriteStyled(styled.MakeText("> ", "red")).
			SetDotToCursor().Buffer())
}

func TestPrompt_ByteOutput(t *testing.T) {
	ttyCtrl, _, cleanup := setupWithRC(`edit:prompt = { put 'bytes> ' }`)
	defer cleanup()

	ttyCtrl.TestBuffer(t,
		bb().WritePlain("bytes> ").SetDotToCursor().Buffer())
}

func TestPrompt_NotifiesInvalidValueOutput(t *testing.T) {
	ttyCtrl, _, cleanup := setupWithRC(`edit:prompt = { put good [bad] good2 }`)
	defer cleanup()

	ttyCtrl.TestBuffer(t,
		bb().WritePlain("goodgood2").SetDotToCursor().Buffer())
	ttyCtrl.TestNotesBuffer(t, bb().
		WritePlain("invalid output type from prompt: list").Buffer())
}

func TestPrompt_NotifiesException(t *testing.T) {
	ttyCtrl, _, cleanup := setupWithRC(`edit:prompt = { fail ERROR }`)
	defer cleanup()

	ttyCtrl.TestNotesBuffer(t, bb().
		WritePlain("prompt function error: ERROR").Buffer())
}

func TestRPrompt(t *testing.T) {
	ttyCtrl, _, cleanup := setupWithRC(`edit:rprompt = { put 'RRR' }`)
	defer cleanup()

	ttyCtrl.TestBuffer(t,
		bb().WritePlain("~> ").SetDotToCursor().
			WritePlain(strings.Repeat(" ", testTTYWidth-6)+"RRR").Buffer())
}

func TestPromptEagerness(t *testing.T) {
	ttyCtrl, _, cleanup := setupWithRC(
		`i = 0`,
		`edit:prompt = { i = (+ $i 1); put $i'> ' }`,
		`edit:-prompt-eagerness = 10`)
	defer cleanup()

	wantBuf1 := bb().WritePlain("1> ").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf1)
	// With eagerness = 10, any key press will cause the prompt to be
	// recomputed.
	ttyCtrl.Inject(term.K(ui.Backspace))
	wantBuf2 := bb().WritePlain("2> ").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf2)
}

func TestPromptStaleThreshold(t *testing.T) {
	ttyCtrl, ev, cleanup := setupWithRC(
		`pipe = (pipe)`,
		`edit:prompt = { nop (slurp < $pipe); put '> ' }`,
		`edit:prompt-stale-threshold = 0.05`)
	defer cleanup()

	wantBufStale := bb().
		WriteStyled(styled.MakeText("???> ", "inverse")).SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBufStale)

	evalf(ev, `pwclose $pipe`)
	wantBufFresh := bb().WritePlain("> ").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBufFresh)
	evalf(ev, `prclose $pipe`)
}

func TestPromptStaleTransform(t *testing.T) {
	ttyCtrl, ev, cleanup := setupWithRC(
		`pipe = (pipe)`,
		`edit:prompt = { nop (slurp < $pipe); put '> ' }`,
		`edit:prompt-stale-threshold = 0.05`,
		`edit:prompt-stale-transform = [a]{ put S; put $a; put S }`)
	defer cleanup()

	wantBufStale := bb().
		WriteStyled(styled.Plain("S???> S")).SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBufStale)
	evalf(ev, `pwclose $pipe`)
	evalf(ev, `prclose $pipe`)
}

func TestRPromptPersistent_True(t *testing.T) {
	wantBufFinal := bb().
		WritePlain("~> " + strings.Repeat(" ", testTTYWidth-6) + "RRR").
		Newline().SetDotToCursor().
		Buffer()
	testRPromptPersistent(t, `edit:rprompt-persistent = $true`, wantBufFinal)
}

func TestRPromptPersistent_False(t *testing.T) {
	wantBufFinal := bb().
		WritePlain("~> "). // no rprompt
		Newline().SetDotToCursor().
		Buffer()
	testRPromptPersistent(t, `edit:rprompt-persistent = $false`, wantBufFinal)
}

func testRPromptPersistent(t *testing.T, code string, wantBufFinal *ui.Buffer) {
	ttyCtrl, _, cleanup := setupWithRC(`edit:rprompt = { put RRR }`, code)
	defer cleanup()

	// Make sure that the UI has stablized before hitting Enter.
	wantBufStable := bb().
		WritePlain("~> ").SetDotToCursor().
		WritePlain(strings.Repeat(" ", testTTYWidth-6) + "RRR").
		Buffer()
	ttyCtrl.TestBuffer(t, wantBufStable)
	ttyCtrl.Inject(term.K('\n'))

	ttyCtrl.TestBuffer(t, wantBufFinal)
}

func TestDefaultPromptForNonRoot(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setupUnstarted()
	defer cleanup()
	ev.Global["f"] = vars.NewReadOnly(getDefaultPrompt(false))
	evalf(ev, `edit:prompt = $f`)

	_, _, stop := start(ed)
	defer stop()

	wantBuf := bb().WritePlain("~> ").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
}

func TestDefaultPromptForRoot(t *testing.T) {
	ed, ttyCtrl, ev, cleanup := setupUnstarted()
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
	ed, ttyCtrl, ev, cleanup := setupUnstarted()
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

func setupWithRC(codes ...string) (cli.TTYCtrl, *eval.Evaler, func()) {
	ed, ttyCtrl, ev, cleanup := setupUnstarted()
	for _, code := range codes {
		evalf(ev, `%s`, code)
	}
	_, _, stop := start(ed)
	return ttyCtrl, ev, func() {
		stop()
		cleanup()
	}
}
