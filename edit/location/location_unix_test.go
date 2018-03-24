// +build !windows,!plan9

package location

import (
	"testing"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/tt"
)

var workspaces = vals.MakeMapFromKV(
	// Pattern is always anchored at beginning; this won't match anything
	"bad", "bad",
	// This is a normal pattern.
	"linux", "/src/linux/[^/]+",
	// Pattern may match a trailing /, in which case it only matches subdirs
	"bsd", "/src/bsd/[^/]+/",
)

func TestMatchWorkspace(t *testing.T) {
	var nilWsInfo *wsInfo

	tt.Test(t, tt.Fn("matchWorkspace", matchWorkspace), tt.Table{
		tt.Args("/bad", workspaces).Rets(nilWsInfo),
		tt.Args("/src/linux/ws1", workspaces).Rets(
			&wsInfo{"linux", "/src/linux/ws1"}),
		tt.Args("/src/linux/ws1/dir", workspaces).Rets(
			&wsInfo{"linux", "/src/linux/ws1"}),
		tt.Args("/src/bsd/ws1", workspaces).Rets(nilWsInfo),
		tt.Args("/src/bsd/ws1/dir", workspaces).Rets(
			&wsInfo{"bsd", "/src/bsd/ws1/"}),
	})
}
