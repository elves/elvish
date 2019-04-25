package lastcmd

import (
	"testing"

	"github.com/elves/elvish/newedit/listing"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

var Args = tt.Args

func TestLastCmdItemsGetter_ShowAll(t *testing.T) {
	g := itemsGetter("put hello elvish", []string{"put", "hello", "elvish"})

	tt.Test(t, tt.Fn("itemsGetter", g), tt.Table{
		// Empty filter; show everything
		Args("").Rets(listing.MatchItems(
			styled.Plain("    put hello elvish"),
			styled.Plain("  0 put"),
			styled.Plain("  1 hello"),
			styled.Plain("  2 elvish"),
		)),
		// Filter = "-", show the individual words with their negative indicies
		Args("-").Rets(listing.MatchItems(
			styled.Plain(" -3 put"),
			styled.Plain(" -2 hello"),
			styled.Plain(" -1 elvish"),
		)),
	})
}

func TestLastCmdItemsGetter_PrefixMatchIndex(t *testing.T) {
	g := itemsGetter("put 1 2 3 4 5 6 7 8 9 10 11 12", []string{
		"put", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"})

	tt.Test(t, tt.Fn("itemsGetter", g), tt.Table{
		Args("1").Rets(listing.MatchItems(
			styled.Plain("  1 1"),
			styled.Plain(" 10 10"),
			styled.Plain(" 11 11"),
			styled.Plain(" 12 12"),
		)),
		Args("-1").Rets(listing.MatchItems(
			styled.Plain("-13 put"),
			styled.Plain("-12 1"),
			styled.Plain("-11 2"),
			styled.Plain("-10 3"),
			styled.Plain(" -1 12"),
		)),
		// Only match prefix; 10 should be missing
		Args("0").Rets(listing.MatchItems(
			styled.Plain("  0 put"),
		)),
	})
}
