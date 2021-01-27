package edit

import (
	"fmt"
	"strings"
	"testing"

	"src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/ui"
)

func TestPrompt_ValueOutput(t *testing.T) {
	f := setup(rc(`edit:prompt = { put '#'; float64 13; styled '> ' red }`))
	defer f.Cleanup()

	f.TestTTY(t,
		"#13> ", Styles,
		"   !!", term.DotHere)
}

func TestPrompt_ByteOutput(t *testing.T) {
	f := setup(rc(`edit:prompt = { print 'bytes> ' }`))
	defer f.Cleanup()

	f.TestTTY(t, "bytes> ", term.DotHere)
}

func TestPrompt_ParsesSGRInByteOutput(t *testing.T) {
	f := setup(rc(`edit:prompt = { print "\033[31mred\033[m> " }`))
	defer f.Cleanup()

	f.TestTTY(t,
		"red> ", Styles,
		"!!!  ", term.DotHere)
}

func TestPrompt_NotifiesInvalidValueOutput(t *testing.T) {
	f := setup(rc(`edit:prompt = { put good [bad] good2 }`))
	defer f.Cleanup()

	f.TestTTY(t, "goodgood2", term.DotHere)
	f.TestTTYNotes(t, "invalid output type from prompt: list")
}

func TestPrompt_NotifiesException(t *testing.T) {
	f := setup(rc(`edit:prompt = { fail ERROR }`))
	defer f.Cleanup()

	f.TestTTYNotes(t,
		"[prompt error] ERROR\n",
		`see stack trace with "show $edit:exceptions[0]"`)
	evals(f.Evaler, `excs = (count $edit:exceptions)`)
	testGlobal(t, f.Evaler, "excs", "1")
}

func TestRPrompt(t *testing.T) {
	f := setup(rc(`edit:rprompt = { put 'RRR' }`))
	defer f.Cleanup()

	f.TestTTY(t, "~> ", term.DotHere,
		strings.Repeat(" ", clitest.FakeTTYWidth-6)+"RRR")
}

func TestPromptEagerness(t *testing.T) {
	f := setup(rc(
		`i = 0`,
		`edit:prompt = { i = (+ $i 1); put $i'> ' }`,
		`edit:-prompt-eagerness = 10`))
	defer f.Cleanup()

	f.TestTTY(t, "1> ", term.DotHere)
	// With eagerness = 10, any key press will cause the prompt to be
	// recomputed.
	f.TTYCtrl.Inject(term.K(ui.Backspace))
	f.TestTTY(t, "2> ", term.DotHere)
}

func TestPromptStaleThreshold(t *testing.T) {
	f := setup(rc(
		`pipe = (pipe)`,
		`edit:prompt = { nop (slurp < $pipe); put '> ' }`,
		`edit:prompt-stale-threshold = `+scaledMsAsSec(50)))
	defer f.Cleanup()

	f.TestTTY(t,
		"???> ", Styles,
		"+++++", term.DotHere)

	evals(f.Evaler, `pwclose $pipe`)
	f.TestTTY(t, "> ", term.DotHere)
	evals(f.Evaler, `prclose $pipe`)
}

func TestPromptStaleTransform(t *testing.T) {
	f := setup(rc(
		`pipe = (pipe)`,
		`edit:prompt = { nop (slurp < $pipe); put '> ' }`,
		`edit:prompt-stale-threshold = `+scaledMsAsSec(50),
		`edit:prompt-stale-transform = [a]{ put S; put $a; put S }`))
	defer f.Cleanup()

	f.TestTTY(t, "S???> S", term.DotHere)
	evals(f.Evaler, `pwclose $pipe`)
	evals(f.Evaler, `prclose $pipe`)
}

func TestPromptStaleTransform_Exception(t *testing.T) {
	f := setup(rc(
		`pipe = (pipe)`,
		`edit:prompt = { nop (slurp < $pipe); put '> ' }`,
		`edit:prompt-stale-threshold = `+scaledMsAsSec(50),
		`edit:prompt-stale-transform = [_]{ fail ERROR }`))
	defer f.Cleanup()

	f.TestTTYNotes(t,
		"[prompt stale transform error] ERROR\n",
		`see stack trace with "show $edit:exceptions[0]"`)
	evals(f.Evaler, `excs = (count $edit:exceptions)`)
	testGlobal(t, f.Evaler, "excs", "1")
}

func TestRPromptPersistent_True(t *testing.T) {
	testRPromptPersistent(t, `edit:rprompt-persistent = $true`,
		"~> "+strings.Repeat(" ", clitest.FakeTTYWidth-6)+"RRR",
		"\n", term.DotHere,
	)
}

func TestRPromptPersistent_False(t *testing.T) {
	testRPromptPersistent(t, `edit:rprompt-persistent = $false`,
		"~> ", // no rprompt
		"\n", term.DotHere,
	)
}

func testRPromptPersistent(t *testing.T, code string, finalBuf ...interface{}) {
	f := setup(rc(`edit:rprompt = { put RRR }`, code))
	defer f.Cleanup()

	// Make sure that the UI has stabilized before hitting Enter.
	f.TestTTY(t,
		"~> ", term.DotHere,
		strings.Repeat(" ", clitest.FakeTTYWidth-6), "RRR",
	)

	f.TTYCtrl.Inject(term.K('\n'))
	f.TestTTY(t, finalBuf...)
}

func TestDefaultPromptForNonRoot(t *testing.T) {
	f := setup(assign("edit:prompt", getDefaultPrompt(false)))
	defer f.Cleanup()

	f.TestTTY(t, "~> ", term.DotHere)
}

func TestDefaultPromptForRoot(t *testing.T) {
	f := setup(assign("edit:prompt", getDefaultPrompt(true)))
	defer f.Cleanup()

	f.TestTTY(t,
		"~# ", Styles,
		" !!", term.DotHere)
}

func TestDefaultRPrompt(t *testing.T) {
	f := setup(assign("edit:rprompt", getDefaultRPrompt("elf", "host")))
	defer f.Cleanup()

	f.TestTTY(t,
		"~> ", term.DotHere, strings.Repeat(" ", 39),
		"elf@host", Styles,
		"++++++++")
}

func scaledMsAsSec(ms int) string {
	return fmt.Sprint(testutil.ScaledMs(ms).Seconds())
}
