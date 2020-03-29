// +build !windows,!plan9,!js

package eval

import (
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

func TestExternalCmdExit_Error(t *testing.T) {
	tt.Test(t, tt.Fn("Error", error.Error), tt.Table{
		tt.Args(ExternalCmdExit{0, "ls", 100}).Rets("ls exited with 0"),
	})
}
