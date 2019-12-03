package apptest

import (
	"testing"

	"github.com/elves/elvish/cli"
)

func TestFixture(t *testing.T) {
	f := Setup(
		WithSpec(func(spec *cli.AppSpec) {
			spec.CodeAreaState.Buffer.Content = "test"
		}),
		WithTTY(func(tty cli.TTYCtrl) {
			tty.SetSize(99, 100)
		}),
	)
	defer f.Stop()

	// Verify that the functions passed to Setup have taken effect.
	if cli.CodeBuffer(f.App).Content != "test" {
		t.Errorf("WithSpec did not work")
	}
	// TODO: Verify the WithTTY function too.

	f.App.CommitCode()
	if code, err := f.Wait(); code != "test" || err != nil {
		t.Errorf("Wait returned %q, %v", code, err)
	}
}
