package edit

import (
	"testing"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/store"
	"src.elv.sh/pkg/ui"
)

func TestBeforeReadline(t *testing.T) {
	f := setup(rc(
		`called = 0`,
		`edit:before-readline = [ { called = (+ $called 1) } ]`))
	defer f.Cleanup()

	// Wait for UI to stabilize so that we can be sure that before-readline hooks
	// have been called.
	f.TestTTY(t, "~> ", term.DotHere)

	testGlobal(t, f.Evaler, "called", 1)
}

func TestAfterReadline(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	evals(f.Evaler,
		`called = 0`,
		`called-with = ''`,
		`edit:after-readline = [
	             [code]{ called = (+ $called 1); called-with = $code } ]`)

	// Wait for UI to stabilize so that we can be sure that after-readline hooks
	// are *not* called.
	f.TestTTY(t, "~> ", term.DotHere)
	testGlobal(t, f.Evaler, "called", "0")

	// Input "test code", press Enter and wait until the editor is done.
	feedInput(f.TTYCtrl, "test code\n")
	f.Wait()

	testGlobals(t, f.Evaler, map[string]interface{}{
		"called":      1,
		"called-with": "test code",
	})
}

func TestAddCmdFilters(t *testing.T) {
	cases := []struct {
		name        string
		rc          string
		input       string
		wantHistory []store.Cmd
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
			wantHistory: []store.Cmd{store.Cmd{Text: "echo", Seq: 1}},
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
			wantHistory: []store.Cmd{store.Cmd{Text: "echo", Seq: 1}},
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

func TestAddCmdFilters_SkipsRemainingOnFalse(t *testing.T) {
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

func TestGlobalBindings(t *testing.T) {
	f := setup(rc(
		`var called = $false`,
		`edit:global-binding[Ctrl-X] = { set called = $true }`,
	))
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K('X', ui.Ctrl))
	f.TTYCtrl.Inject(term.K(ui.Enter))
	f.Wait()

	testGlobal(t, f.Evaler, "called", true)
}
