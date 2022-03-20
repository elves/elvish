package edit

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/ui"
)

func TestPrompt_ValueOutput(t *testing.T) {
	f := setup(t, rc(`set edit:prompt = { put '#'; num 13; styled '> ' red }`))

	f.TestTTY(t,
		"#13> ", Styles,
		"   !!", term.DotHere)
}

func TestPrompt_ByteOutput(t *testing.T) {
	f := setup(t, rc(`set edit:prompt = { print 'bytes> ' }`))

	f.TestTTY(t, "bytes> ", term.DotHere)
}

func TestPrompt_ParsesSGRInByteOutput(t *testing.T) {
	f := setup(t, rc(`set edit:prompt = { print "\033[31mred\033[m> " }`))

	f.TestTTY(t,
		"red> ", Styles,
		"!!!  ", term.DotHere)
}

func TestPrompt_NotifiesInvalidValueOutput(t *testing.T) {
	f := setup(t, rc(`set edit:prompt = { put good [bad] good2 }`))

	f.TestTTY(t, "goodgood2", term.DotHere)
	f.TestTTYNotes(t, "invalid output type from prompt: list")
}

func TestPrompt_NotifiesException(t *testing.T) {
	f := setup(t, rc(`set edit:prompt = { fail ERROR }`))

	f.TestTTYNotes(t,
		"[prompt error] ERROR\n",
		`see stack trace with "show $edit:exceptions[0]"`)
	evals(f.Evaler, `var excs = (count $edit:exceptions)`)
	testGlobal(t, f.Evaler, "excs", 1)
}

func TestRPrompt(t *testing.T) {
	f := setup(t, rc(`set edit:rprompt = { put 'RRR' }`))

	f.TestTTY(t, "~> ", term.DotHere,
		strings.Repeat(" ", clitest.FakeTTYWidth-6)+"RRR")
}

func TestPromptEagerness(t *testing.T) {
	f := setup(t, rc(
		`var i = 0`,
		`set edit:prompt = { set i = (+ $i 1); put $i'> ' }`,
		`set edit:-prompt-eagerness = 10`))

	f.TestTTY(t, "1> ", term.DotHere)
	// With eagerness = 10, any key press will cause the prompt to be
	// recomputed.
	f.TTYCtrl.Inject(term.K(ui.Backspace))
	f.TestTTY(t, "2> ", term.DotHere)
}

func TestPromptStaleThreshold(t *testing.T) {
	f := setup(t, rc(
		`var pipe = (file:pipe)`,
		`set edit:prompt = { nop (slurp < $pipe); put '> ' }`,
		`set edit:prompt-stale-threshold = `+scaledMsAsSec(50)))

	f.TestTTY(t,
		"???> ", Styles,
		"+++++", term.DotHere)

	evals(f.Evaler, `file:close $pipe[w]`)
	f.TestTTY(t, "> ", term.DotHere)
	evals(f.Evaler, `file:close $pipe[r]`)
}

func TestPromptStaleTransform(t *testing.T) {
	f := setup(t, rc(
		`var pipe = (file:pipe)`,
		`set edit:prompt = { nop (slurp < $pipe); put '> ' }`,
		`set edit:prompt-stale-threshold = `+scaledMsAsSec(50),
		`set edit:prompt-stale-transform = {|a| put S; put $a; put S }`))

	f.TestTTY(t, "S???> S", term.DotHere)
	evals(f.Evaler, `file:close $pipe[w]`)
	evals(f.Evaler, `file:close $pipe[r]`)
}

func TestPromptStaleTransform_Exception(t *testing.T) {
	f := setup(t, rc(
		`var pipe = (file:pipe)`,
		`set edit:prompt = { nop (slurp < $pipe); put '> ' }`,
		`set edit:prompt-stale-threshold = `+scaledMsAsSec(50),
		`set edit:prompt-stale-transform = {|_| fail ERROR }`))

	f.TestTTYNotes(t,
		"[prompt stale transform error] ERROR\n",
		`see stack trace with "show $edit:exceptions[0]"`)
	evals(f.Evaler, `var excs = (count $edit:exceptions)`)
	testGlobal(t, f.Evaler, "excs", 1)
}

func TestRPromptPersistent_True(t *testing.T) {
	testRPromptPersistent(t, `set edit:rprompt-persistent = $true`,
		"~> "+strings.Repeat(" ", clitest.FakeTTYWidth-6)+"RRR",
		"\n", term.DotHere,
	)
}

func TestRPromptPersistent_False(t *testing.T) {
	testRPromptPersistent(t, `set edit:rprompt-persistent = $false`,
		"~> ", // no rprompt
		"\n", term.DotHere,
	)
}

func testRPromptPersistent(t *testing.T, code string, finalBuf ...any) {
	f := setup(t, rc(`set edit:rprompt = { put RRR }`, code))

	// Make sure that the UI has stabilized before hitting Enter.
	f.TestTTY(t,
		"~> ", term.DotHere,
		strings.Repeat(" ", clitest.FakeTTYWidth-6), "RRR",
	)

	f.TTYCtrl.Inject(term.K('\n'))
	f.TestTTY(t, finalBuf...)
}

func TestDefaultPromptForNonRoot(t *testing.T) {
	f := setup(t, assign("edit:prompt", getDefaultPrompt(false)))

	f.TestTTY(t, "~> ", term.DotHere)
}

func TestDefaultPromptForRoot(t *testing.T) {
	f := setup(t, assign("edit:prompt", getDefaultPrompt(true)))

	f.TestTTY(t,
		"~# ", Styles,
		" !!", term.DotHere)
}

func TestDefaultRPrompt(t *testing.T) {
	f := setup(t, assign("edit:rprompt", getDefaultRPrompt("elf", "host")))

	f.TestTTY(t,
		"~> ", term.DotHere, strings.Repeat(" ", 39),
		"elf@host", Styles,
		"++++++++")
}

func scaledMsAsSec(ms int) string {
	return fmt.Sprint(testutil.Scaled(time.Duration(ms) * time.Millisecond).Seconds())
}
