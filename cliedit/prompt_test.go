package cliedit

import (
	"strings"
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/styled"
)

func TestPrompt_ValueOutput(t *testing.T) {
	f := setupWithRC(`edit:prompt = { put 'val'; styled '> ' red }`)
	defer f.Cleanup()

	f.TTYCtrl.TestBuffer(t,
		bb().WritePlain("val").WriteStyled(styled.MakeText("> ", "red")).
			SetDotHere().Buffer())
}

func TestPrompt_ByteOutput(t *testing.T) {
	f := setupWithRC(`edit:prompt = { put 'bytes> ' }`)
	defer f.Cleanup()

	f.TTYCtrl.TestBuffer(t,
		bb().WritePlain("bytes> ").SetDotHere().Buffer())
}

func TestPrompt_NotifiesInvalidValueOutput(t *testing.T) {
	f := setupWithRC(`edit:prompt = { put good [bad] good2 }`)
	defer f.Cleanup()

	f.TTYCtrl.TestBuffer(t,
		bb().WritePlain("goodgood2").SetDotHere().Buffer())
	f.TTYCtrl.TestNotesBuffer(t, bb().
		WritePlain("invalid output type from prompt: list").Buffer())
}

func TestPrompt_NotifiesException(t *testing.T) {
	f := setupWithRC(`edit:prompt = { fail ERROR }`)
	defer f.Cleanup()

	f.TTYCtrl.TestNotesBuffer(t, bb().
		WritePlain("prompt function error: ERROR").Buffer())
}

func TestRPrompt(t *testing.T) {
	f := setupWithRC(`edit:rprompt = { put 'RRR' }`)
	defer f.Cleanup()

	f.TTYCtrl.TestBuffer(t,
		bb().WritePlain("~> ").SetDotHere().
			WritePlain(strings.Repeat(" ", testTTYWidth-6)+"RRR").Buffer())
}

func TestPromptEagerness(t *testing.T) {
	f := setupWithRC(
		`i = 0`,
		`edit:prompt = { i = (+ $i 1); put $i'> ' }`,
		`edit:-prompt-eagerness = 10`)
	defer f.Cleanup()

	wantBuf1 := bb().WritePlain("1> ").SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf1)
	// With eagerness = 10, any key press will cause the prompt to be
	// recomputed.
	f.TTYCtrl.Inject(term.K(ui.Backspace))
	wantBuf2 := bb().WritePlain("2> ").SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf2)
}

func TestPromptStaleThreshold(t *testing.T) {
	f := setupWithRC(
		`pipe = (pipe)`,
		`edit:prompt = { nop (slurp < $pipe); put '> ' }`,
		`edit:prompt-stale-threshold = 0.05`)
	defer f.Cleanup()

	wantBufStale := bb().
		WriteStyled(styled.MakeText("???> ", "inverse")).SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufStale)

	evals(f.Evaler, `pwclose $pipe`)
	wantBufFresh := bb().WritePlain("> ").SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufFresh)
	evals(f.Evaler, `prclose $pipe`)
}

func TestPromptStaleTransform(t *testing.T) {
	f := setupWithRC(
		`pipe = (pipe)`,
		`edit:prompt = { nop (slurp < $pipe); put '> ' }`,
		`edit:prompt-stale-threshold = 0.05`,
		`edit:prompt-stale-transform = [a]{ put S; put $a; put S }`)
	defer f.Cleanup()

	wantBufStale := bb().
		WriteStyled(styled.Plain("S???> S")).SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufStale)
	evals(f.Evaler, `pwclose $pipe`)
	evals(f.Evaler, `prclose $pipe`)
}

func TestRPromptPersistent_True(t *testing.T) {
	wantBufFinal := bb().
		WritePlain("~> " + strings.Repeat(" ", testTTYWidth-6) + "RRR").
		Newline().SetDotHere().
		Buffer()
	testRPromptPersistent(t, `edit:rprompt-persistent = $true`, wantBufFinal)
}

func TestRPromptPersistent_False(t *testing.T) {
	wantBufFinal := bb().
		WritePlain("~> "). // no rprompt
		Newline().SetDotHere().
		Buffer()
	testRPromptPersistent(t, `edit:rprompt-persistent = $false`, wantBufFinal)
}

func testRPromptPersistent(t *testing.T, code string, wantBufFinal *ui.Buffer) {
	f := setupWithRC(`edit:rprompt = { put RRR }`, code)
	defer f.Cleanup()

	// Make sure that the UI has stablized before hitting Enter.
	wantBufStable := bb().
		WritePlain("~> ").SetDotHere().
		WritePlain(strings.Repeat(" ", testTTYWidth-6) + "RRR").
		Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufStable)
	f.TTYCtrl.Inject(term.K('\n'))

	f.TTYCtrl.TestBuffer(t, wantBufFinal)
}

func TestDefaultPromptForNonRoot(t *testing.T) {
	f := setupWithOpt(setupOpt{Unstarted: true})
	defer f.Cleanup()
	f.Evaler.Global["f"] = vars.NewReadOnly(getDefaultPrompt(false))
	evals(f.Evaler, `edit:prompt = $f`)

	f.Start()

	wantBuf := bb().WritePlain("~> ").SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestDefaultPromptForRoot(t *testing.T) {
	f := setupWithOpt(setupOpt{Unstarted: true})
	defer f.Cleanup()
	f.Evaler.Global["f"] = vars.NewReadOnly(getDefaultPrompt(true))
	evals(f.Evaler, `edit:prompt = $f`)

	f.Start()

	wantBuf := bb().WritePlain("~").
		WriteStyled(styled.MakeText("# ", "red")).SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestDefaultRPrompt(t *testing.T) {
	f := setupWithOpt(setupOpt{Unstarted: true})
	defer f.Cleanup()
	f.Evaler.Global["f"] = vars.NewReadOnly(getDefaultRPrompt("elf", "host"))
	evals(f.Evaler, `edit:rprompt = $f`)

	f.Start()

	wantBuf := bb().WritePlain("~> ").SetDotHere().
		WritePlain(strings.Repeat(" ", 49)).
		WriteStyled(styled.MakeText("elf@host", "inverse")).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}
