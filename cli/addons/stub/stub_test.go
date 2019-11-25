package stub

import (
	"testing"
	"time"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
)

func TestRendering(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	Start(app, Config{Name: " STUB "})
	modeline := layout.ModeLine(" STUB ", false)
	ttyCtrl.TestBuffer(t,
		bb().SetDotHere().Newline().WriteStyled(modeline).Buffer())
}

func TestFocus(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	Start(app, Config{Name: " STUB ", Focus: true})
	modeline := layout.ModeLine(" STUB ", false)
	ttyCtrl.TestBuffer(t,
		bb().Newline().WriteStyled(modeline).SetDotHere().Buffer())
}

func TestHandling(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	bindingCalled := make(chan bool)
	Start(app, Config{
		Binding: el.MapHandler{term.K('a'): func() { bindingCalled <- true }},
	})

	ttyCtrl.Inject(term.K('a'))
	select {
	case <-bindingCalled:
		// OK
	case <-time.After(time.Second):
		t.Errorf("Handler not called after 1s")
	}
}

func setup() (cli.App, cli.TTYCtrl, func()) {
	tty, ttyCtrl := cli.NewFakeTTY()
	ttyCtrl.SetSize(24, 40)
	app := cli.NewApp(cli.AppSpec{TTY: tty})
	codeCh, _ := cli.ReadCodeAsync(app)
	return app, ttyCtrl, func() {
		app.CommitEOF()
		<-codeCh
	}
}

func bb() *term.BufferBuilder { return term.NewBufferBuilder(40) }
