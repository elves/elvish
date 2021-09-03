package modes

import (
	"errors"
	"testing"

	"src.elv.sh/pkg/cli"
	. "src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/ui"
)

func setupStartedInstant(t *testing.T) *Fixture {
	f := Setup()
	w, err := NewInstant(f.App, InstantSpec{
		Execute: func(code string) ([]string, error) {
			var err error
			if code == "!" {
				err = errors.New("error")
			}
			return []string{"result of", code}, err
		},
	})
	startMode(f.App, w, err)
	f.TestTTY(t,
		term.DotHere, "\n",
		" INSTANT \n", Styles,
		"*********",
		"result of\n",
		"",
	)
	return f
}

func TestInstant_ShowsResult(t *testing.T) {
	f := setupStartedInstant(t)
	defer f.Stop()

	f.TTY.Inject(term.K('a'))
	bufA := f.MakeBuffer(
		"a", term.DotHere, "\n",
		" INSTANT \n", Styles,
		"*********",
		"result of\n",
		"a",
	)
	f.TTY.TestBuffer(t, bufA)

	f.TTY.Inject(term.K(ui.Right))
	f.TTY.TestBuffer(t, bufA)
}

func TestInstant_ShowsError(t *testing.T) {
	f := setupStartedInstant(t)
	defer f.Stop()

	f.TTY.Inject(term.K('!'))
	f.TestTTY(t,
		"!", term.DotHere, "\n",
		" INSTANT \n", Styles,
		"*********",
		// Error shown.
		"error\n", Styles,
		"!!!!!",
		// Buffer not updated.
		"result of\n",
		"",
	)
}

func TestNewInstant_NoExecutor(t *testing.T) {
	f := Setup()
	_, err := NewInstant(f.App, InstantSpec{})
	if err != errExecutorIsRequired {
		t.Error("expect errExecutorIsRequired")
	}
}

func TestNewInstant_FocusedWidgetNotCodeArea(t *testing.T) {
	testFocusedWidgetNotCodeArea(t, func(app cli.App) error {
		_, err := NewInstant(app, InstantSpec{
			Execute: func(string) ([]string, error) { return nil, nil }})
		return err
	})
}
