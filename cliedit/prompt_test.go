package cliedit

import (
	"strings"
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/ui"
)

func TestPrompt_ValueOutput(t *testing.T) {
	f := setup(rc(`edit:prompt = { put 'val'; styled '> ' red }`))
	defer f.Cleanup()

	f.TTYCtrl.TestBuffer(t,
		bb().Write("val").Write("> ", ui.Red).
			SetDotHere().Buffer())
}

func TestPrompt_ByteOutput(t *testing.T) {
	f := setup(rc(`edit:prompt = { put 'bytes> ' }`))
	defer f.Cleanup()

	f.TTYCtrl.TestBuffer(t,
		bb().Write("bytes> ").SetDotHere().Buffer())
}

func TestPrompt_NotifiesInvalidValueOutput(t *testing.T) {
	f := setup(rc(`edit:prompt = { put good [bad] good2 }`))
	defer f.Cleanup()

	f.TTYCtrl.TestBuffer(t,
		bb().Write("goodgood2").SetDotHere().Buffer())
	f.TTYCtrl.TestNotesBuffer(t, bb().
		Write("invalid output type from prompt: list").Buffer())
}

func TestPrompt_NotifiesException(t *testing.T) {
	f := setup(rc(`edit:prompt = { fail ERROR }`))
	defer f.Cleanup()

	f.TTYCtrl.TestNotesBuffer(t, bb().
		Write("prompt function error: ERROR").Buffer())
}

func TestRPrompt(t *testing.T) {
	f := setup(rc(`edit:rprompt = { put 'RRR' }`))
	defer f.Cleanup()

	f.TTYCtrl.TestBuffer(t,
		bb().Write("~> ").SetDotHere().
			Write(strings.Repeat(" ", testTTYWidth-6)+"RRR").Buffer())
}

func TestPromptEagerness(t *testing.T) {
	f := setup(rc(
		`i = 0`,
		`edit:prompt = { i = (+ $i 1); put $i'> ' }`,
		`edit:-prompt-eagerness = 10`))
	defer f.Cleanup()

	wantBuf1 := bb().Write("1> ").SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf1)
	// With eagerness = 10, any key press will cause the prompt to be
	// recomputed.
	f.TTYCtrl.Inject(term.K(ui.Backspace))
	wantBuf2 := bb().Write("2> ").SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf2)
}

func TestPromptStaleThreshold(t *testing.T) {
	f := setup(rc(
		`pipe = (pipe)`,
		`edit:prompt = { nop (slurp < $pipe); put '> ' }`,
		`edit:prompt-stale-threshold = 0.05`))
	defer f.Cleanup()

	wantBufStale := bb().
		Write("???> ", ui.Inverse).SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufStale)

	evals(f.Evaler, `pwclose $pipe`)
	wantBufFresh := bb().Write("> ").SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufFresh)
	evals(f.Evaler, `prclose $pipe`)
}

func TestPromptStaleTransform(t *testing.T) {
	f := setup(rc(
		`pipe = (pipe)`,
		`edit:prompt = { nop (slurp < $pipe); put '> ' }`,
		`edit:prompt-stale-threshold = 0.05`,
		`edit:prompt-stale-transform = [a]{ put S; put $a; put S }`))
	defer f.Cleanup()

	wantBufStale := bb().
		WriteStyled(ui.T("S???> S")).SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufStale)
	evals(f.Evaler, `pwclose $pipe`)
	evals(f.Evaler, `prclose $pipe`)
}

func TestRPromptPersistent_True(t *testing.T) {
	wantBufFinal := bb().
		Write("~> " + strings.Repeat(" ", testTTYWidth-6) + "RRR").
		Newline().SetDotHere().
		Buffer()
	testRPromptPersistent(t, `edit:rprompt-persistent = $true`, wantBufFinal)
}

func TestRPromptPersistent_False(t *testing.T) {
	wantBufFinal := bb().
		Write("~> "). // no rprompt
		Newline().SetDotHere().
		Buffer()
	testRPromptPersistent(t, `edit:rprompt-persistent = $false`, wantBufFinal)
}

func testRPromptPersistent(t *testing.T, code string, wantBufFinal *term.Buffer) {
	f := setup(rc(`edit:rprompt = { put RRR }`, code))
	defer f.Cleanup()

	// Make sure that the UI has stablized before hitting Enter.
	wantBufStable := bb().
		Write("~> ").SetDotHere().
		Write(strings.Repeat(" ", testTTYWidth-6) + "RRR").
		Buffer()
	f.TTYCtrl.TestBuffer(t, wantBufStable)
	f.TTYCtrl.Inject(term.K('\n'))

	f.TTYCtrl.TestBuffer(t, wantBufFinal)
}

func TestDefaultPromptForNonRoot(t *testing.T) {
	f := setup(assign("edit:prompt", getDefaultPrompt(false)))
	defer f.Cleanup()

	wantBuf := bb().Write("~> ").SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestDefaultPromptForRoot(t *testing.T) {
	f := setup(assign("edit:prompt", getDefaultPrompt(true)))
	defer f.Cleanup()

	wantBuf := bb().Write("~").
		Write("# ", ui.Red).SetDotHere().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestDefaultRPrompt(t *testing.T) {
	f := setup(assign("edit:rprompt", getDefaultRPrompt("elf", "host")))
	defer f.Cleanup()

	wantBuf := bb().Write("~> ").SetDotHere().
		Write(strings.Repeat(" ", 49)).
		Write("elf@host", ui.Inverse).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}
