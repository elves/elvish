package histlist

import (
	"testing"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/newedit/listing"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

var Args = tt.Args

func TestGetEntries(t *testing.T) {
	cmds := []histutil.Entry{
		{Text: "put 1", Seq: 1},
		{Text: "echo 2", Seq: 2},
		{Text: "print 3", Seq: 3},
		{Text: "repr 4", Seq: 4},
	}

	tt.Test(t, tt.Fn("getEntries", getItems), tt.Table{
		// Show all commands.
		Args(cmds, "").Rets(listing.MatchItems(
			styled.Plain("   1 put 1"),
			styled.Plain("   2 echo 2"),
			styled.Plain("   3 print 3"),
			styled.Plain("   4 repr 4"),
		)),
		// Filter.
		Args(cmds, "pr").Rets(listing.MatchItems(
			styled.Plain("   3 print 3"),
			styled.Plain("   4 repr 4"),
		)),
	})
}

func TestAccept(t *testing.T) {
	cmds := []histutil.Entry{
		{Text: "put 1", Seq: 1},
		{Text: "echo 2", Seq: 2},
	}
	entries := getItems(cmds, "")
	st := clitypes.State{}

	entries.Accept(0, &st)
	if st.Code() != "put 1" {
		t.Errorf("Accept doesn't insert command")
	}

	entries.Accept(1, &st)
	if st.Code() != "put 1\necho 2" {
		t.Errorf("Accept doesn't insert command with newline")
	}
}
