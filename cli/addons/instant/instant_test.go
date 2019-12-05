package instant

import (
	"errors"
	"testing"

	. "github.com/elves/elvish/cli/apptest"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/ui"
)

var styles = ui.RuneStylesheet{
	'*': ui.Stylings(ui.Bold, ui.LightGray, ui.BgMagenta),
	'!': ui.Red,
}

func setupStarted(t *testing.T) *Fixture {
	f := Setup()
	Start(f.App, Config{
		Execute: func(code string) ([]string, error) {
			var err error
			if code == "!" {
				err = errors.New("error")
			}
			return []string{"result of", code}, err
		},
	})
	f.TestTTY(t,
		term.DotHere, "\n",
		" INSTANT \n", styles,
		"*********",
		"result of\n",
		"",
	)
	return f
}

func TestUpdate(t *testing.T) {
	f := setupStarted(t)
	defer f.Stop()

	f.TTY.Inject(term.K('a'))
	bufA := f.MakeBuffer(
		"a", term.DotHere, "\n",
		" INSTANT \n", styles,
		"*********",
		"result of\n",
		"a",
	)
	f.TTY.TestBuffer(t, bufA)

	f.TTY.Inject(term.K(ui.Right))
	f.TTY.TestBuffer(t, bufA)
}

func TestError(t *testing.T) {
	f := setupStarted(t)
	defer f.Stop()

	f.TTY.Inject(term.K('!'))
	f.TestTTY(t,
		"!", term.DotHere, "\n",
		" INSTANT \n", styles,
		"*********",
		// Error shown.
		"error\n", styles,
		"!!!!!",
		// Buffer not updated.
		"result of\n",
		"",
	)
}

func TestStart_NoExecutor(t *testing.T) {
	f := Setup()
	Start(f.App, Config{})
	f.TestTTYNotes(t, "executor is required")
}
