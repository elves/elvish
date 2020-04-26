package eval

import (
	"testing"

	"github.com/elves/elvish/pkg/util"
)

func TestDeprecate(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()

	TestWithSetup(t, func(ev *Evaler) {
		ev.Global.AddGoFn("", "dep", func(fm *Frame) {
			fm.Deprecate("deprecated")
		})
	},
		That("dep").PrintsStderrWith("deprecated"),
		// Deprecation message is only shown once.
		That("dep 2> tmp.txt; dep").DoesNothing(),
	)
}
