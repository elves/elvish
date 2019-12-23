package stub

import (
	"testing"
	"time"

	. "github.com/elves/elvish/pkg/cli/apptest"
	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/term"
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
		Binding: el.MapHandler{term.K('a'): func() { bindingCalled <- true }},
	})

	f.TTY.Inject(term.K('a'))
	select {
	case <-bindingCalled:
		// OK
	case <-time.After(time.Second):
		t.Errorf("Handler not called after 1s")
	}
}
