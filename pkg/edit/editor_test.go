package edit

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/pkg/store"
)

func TestEditor_AddsHistoryAfterAccepting(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	feedInput(f.TTYCtrl, "echo x\n")
	f.Wait()

	testCommands(t, f.Store, "echo x")
}

func TestEditor_DoesNotAddEmptyCommandToHistory(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	feedInput(f.TTYCtrl, "\n")
	f.Wait()

	testCommands(t, f.Store /* no commands */)
}

func TestEditor_TestAddCmdFilters(t *testing.T) {
	cases := []struct {
		name        string
		rc          string
		input       string
		wantHistory []string
	}{
		// TODO: Enable the following two tests once error output can
		// be tested.
		// {
		// 	name:        "non callable item",
		// 	rc:          "edit:add-cmd-filters = [$false]",
		// 	input:       "echo\n",
		// 	wantHistory: []string{"echo"},
		// },
		// {
		// 	name:        "callback outputs nothing",
		// 	rc:          "edit:add-cmd-filters = [[_]{}]",
		// 	input:       "echo\n",
		// 	wantHistory: []string{"echo"},
		// },
		{
			name:        "callback outputs true",
			rc:          "edit:add-cmd-filters = [[_]{ put $true }]",
			input:       "echo\n",
			wantHistory: []string{"echo"},
		},
		{
			name:        "callback outputs false",
			rc:          "edit:add-cmd-filters = [[_]{ put $false }]",
			input:       "echo\n",
			wantHistory: nil,
		},
		{
			name:        "false-true chain",
			rc:          "edit:add-cmd-filters = [[_]{ put $false } [_]{ put $true }]",
			input:       "echo\n",
			wantHistory: nil,
		},
		{
			name:        "true-false chain",
			rc:          "edit:add-cmd-filters = [[_]{ put $true } [_]{ put $false }]",
			input:       "echo\n",
			wantHistory: nil,
		},
		{
			name:        "positive",
			rc:          "edit:add-cmd-filters = [[cmd]{ ==s $cmd echo }]",
			input:       "echo\n",
			wantHistory: []string{"echo"},
		},
		{
			name:        "negative",
			rc:          "edit:add-cmd-filters = [[cmd]{ ==s $cmd echo }]",
			input:       "echo x\n",
			wantHistory: nil,
		},
		{
			name:        "default value",
			rc:          "",
			input:       " echo\n",
			wantHistory: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := setup(rc(c.rc))
			defer f.Cleanup()

			feedInput(f.TTYCtrl, c.input)
			f.Wait()

			testCommands(t, f.Store, c.wantHistory...)
		})
	}
}

func TestEditor_AddCmdFiltersHasShortCircuit(t *testing.T) {
	f := setup(rc(
		`called = $false`,
		`@edit:add-cmd-filters = [_]{ put $false } [_]{ called = $true; put $true }`,
	))
	defer f.Cleanup()

	feedInput(f.TTYCtrl, "echo\n")
	f.Wait()
	testCommands(t, f.Store)
	testGlobal(t, f.Evaler, "called", false)
}

func testCommands(t *testing.T, store store.Store, wantCmds ...string) {
	t.Helper()
	cmds, err := store.Cmds(0, 1024)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(cmds, wantCmds) {
		t.Errorf("got cmds %v, want %v", cmds, wantCmds)
	}
}
