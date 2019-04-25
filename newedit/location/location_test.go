package location

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/newedit/listing"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

var Args = tt.Args

func TestGetEntries(t *testing.T) {
	dirs := []storedefs.Dir{
		{Path: "/home/elf", Score: 20},
		{Path: "/usr/bin", Score: 10},
	}
	dummyCd := func(string) error { return nil }
	tt.Test(t, tt.Fn("getItems", getItems), tt.Table{
		Args(dirs, "", dummyCd).Rets(listing.MatchItems(
			styled.Unstyled(" 20 /home/elf"),
			styled.Unstyled(" 10 /usr/bin"),
		)),
		Args(dirs, "/usr", dummyCd).Rets(listing.MatchItems(
			styled.Unstyled(" 10 /usr/bin"),
		))})
}

func TestAccept_OK(t *testing.T) {
	dirs := []storedefs.Dir{
		{Path: "/home/elf", Score: 20},
		{Path: "/usr/bin", Score: 10},
	}
	calledDir := ""
	cd := func(dir string) error {
		calledDir = dir
		return nil
	}
	getItems(dirs, "", cd).Accept(0, &clitypes.State{})
	if calledDir != "/home/elf" {
		t.Errorf("Accept did not call cd")
	}
}

func TestAccept_Error(t *testing.T) {
	dirs := []storedefs.Dir{
		{Path: "/home/elf", Score: 20},
		{Path: "/usr/bin", Score: 10},
	}
	cd := func(string) error { return errors.New("cannot cd") }
	state := clitypes.State{}

	getItems(dirs, "", cd).Accept(0, &state)

	wantNotes := []string{"cannot cd"}
	if !reflect.DeepEqual(state.Raw.Notes, wantNotes) {
		t.Errorf("cd errors not added to notes")
	}
}
