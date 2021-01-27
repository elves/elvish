package stub

import (
	"testing"
	"time"

	"src.elv.sh/pkg/cli"
	. "src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/term"
)

func TestRendering(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{Name: " STUB "})
	f.TestTTY(t,
		"", term.DotHere, "\n",
		" STUB ", Styles,
		"******",
	)
}

func TestFocus(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{Name: " STUB ", Focus: true})
	f.TestTTY(t,
		"\n",
		" STUB ", Styles,
		"******", term.DotHere,
	)
}

func TestHandling(t *testing.T) {
	f := Setup()
	defer f.Stop()

	bindingCalled := make(chan bool)
	Start(f.App, Config{
		Binding: cli.MapHandler{term.K('a'): func() { bindingCalled <- true }},
	})

	f.TTY.Inject(term.K('a'))
	select {
	case <-bindingCalled:
		// OK
	case <-time.After(time.Second):
		t.Errorf("Handler not called after 1s")
	}
}
