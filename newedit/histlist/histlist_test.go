package histlist

import (
	"testing"

	"github.com/elves/elvish/newedit/listing"
	"github.com/elves/elvish/newedit/types"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

var Args = tt.Args

var testCmds = []string{}

func TestGetEntries(t *testing.T) {
	cmds := []string{
		"put 1",
		"echo 2",
		"print 3",
		"repr 4",
	}

	tt.Test(t, tt.Fn("getEntries", getEntries), tt.Table{
		// Show all commands.
		Args(cmds, "").Rets(listing.MatchItems(
			styled.Unstyled("   1 put 1"),
			styled.Unstyled("   2 echo 2"),
			styled.Unstyled("   3 print 3"),
			styled.Unstyled("   4 repr 4"),
		)),
		// Filter.
		Args(cmds, "pr").Rets(listing.MatchItems(
			styled.Unstyled("   3 print 3"),
			styled.Unstyled("   4 repr 4"),
		)),
	})
}

func TestAccept(t *testing.T) {
	cmds := []string{
		"put 1",
		"echo 2",
	}
	entries := getEntries(cmds, "")
	st := types.State{}

	entries.Accept(0, &st)
	if st.Code() != "put 1" {
		t.Errorf("Accept doesn't insert command")
	}

	entries.Accept(1, &st)
	if st.Code() != "put 1\necho 2" {
		t.Errorf("Accept doesn't insert command with newline")
	}
}
