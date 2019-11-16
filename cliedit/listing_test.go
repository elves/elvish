package cliedit

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

/*
func TestInitListing_Binding(t *testing.T) {
	// Test that the binding variable in the returned namespace indeed refers to
	// the BindingMap returned.
	_, binding, ns := initListing(&fakeApp{})
	if ns["binding"].Get() != *binding {
		t.Errorf("The binding var in the ns is not the same as the BindingMap")
	}
}
*/

// Smoke tests for individual addons.

func TestHistlistAddon(t *testing.T) {
	f := setupWithOpt(setupOpt{StoreOp: func(s storedefs.Store) {
		s.AddCmd("ls")
		s.AddCmd("echo")
		s.AddCmd("ls")
	}})
	f.TTYCtrl.SetSize(24, 30) // Set width to 30
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K('R', ui.Ctrl))
	wantBuf := bbAddon(" HISTORY (dedup on) ").
		WriteStyled(styled.MarkLines(
			"   1 echo",
			"   2 ls                       ", styles,
			"##############################",
		)).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)

	evals(f.Evaler, `edit:histlist:toggle-dedup`)
	wantBuf = bbAddon(" HISTORY ").
		WriteStyled(styled.MarkLines(
			"   0 ls",
			"   1 echo",
			"   2 ls                       ", styles,
			"##############################",
		)).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)

	evals(f.Evaler, `edit:histlist:toggle-case-sensitivity`)
	wantBuf = bbAddon(" HISTORY (case-insensitive) ").
		WriteStyled(styled.MarkLines(
			"   0 ls",
			"   1 echo",
			"   2 ls                       ", styles,
			"##############################",
		)).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestLastCmdAddon(t *testing.T) {
	f := setupWithOpt(setupOpt{StoreOp: func(s storedefs.Store) {
		s.AddCmd("echo hello world")
	}})
	f.TTYCtrl.SetSize(24, 30) // Set width to 30
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K(',', ui.Alt))
	wantBuf := bbAddon("LASTCMD").
		WriteStyled(styled.MarkLines(
			"    echo hello world          ", styles,
			"##############################",
			"  0 echo",
			"  1 hello",
			"  2 world",
		)).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestLocationAddon(t *testing.T) {
	f := setupWithOpt(setupOpt{StoreOp: func(s storedefs.Store) {
		s.AddDir("/usr/bin", 1)
		s.AddDir("/tmp", 1)
		s.AddDir("/home/elf", 1)
	}})
	f.TTYCtrl.SetSize(24, 30) // Set width to 30
	defer f.Cleanup()

	evals(f.Evaler,
		`edit:location:pinned = [/opt]`,
		`edit:location:hidden = [/tmp]`)
	f.TTYCtrl.Inject(term.K('L', ui.Ctrl))

	wantBuf := bbAddon("LOCATION").
		WriteStyled(styled.MarkLines(
			"  * /opt                      ", styles,
			"##############################",
			" 10 /home/elf",
			" 10 /usr/bin",
		)).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestLocationAddon_Workspace(t *testing.T) {
	f := setupWithOpt(setupOpt{StoreOp: func(s storedefs.Store) {
		s.AddDir("/usr/bin", 1)
		s.AddDir("ws/bin", 1)
		s.AddDir("other-ws/bin", 1)
	}})
	defer f.Cleanup()
	util.ApplyDir(
		util.Dir{
			"ws1": util.Dir{
				"bin": util.Dir{},
				"tmp": util.Dir{}}})
	err := os.Chdir("ws1/tmp")
	if err != nil {
		t.Skip("chdir:", err)
	}
	f.TTYCtrl.SetSize(24, 30) // Set width to 30

	evals(f.Evaler,
		`edit:location:workspaces = [&ws=$E:HOME/ws.]`)

	f.TTYCtrl.Inject(term.K('L', ui.Ctrl))
	wantBuf := ui.NewBufferBuilder(30).
		WritePlain("~/ws1/tmp> ").Newline().
		WriteStyled(layout.ModeLine("LOCATION", true)).SetDotToCursor().Newline().
		WriteStyled(styled.MarkLines(
			" 10 ws/bin                    ", styles,
			"##############################",
			" 10 /usr/bin",
		)).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)

	f.TTYCtrl.Inject(term.K(ui.Enter))
	wantBuf = ui.NewBufferBuilder(30).
		WritePlain("~/ws1/bin> ").SetDotToCursor().Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestLocation_AddDir(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	util.ApplyDir(
		util.Dir{
			"bin": util.Dir{},
			"ws1": util.Dir{
				"bin": util.Dir{}}})
	evals(f.Evaler, `edit:location:workspaces = [&ws=$E:HOME/ws.]`)

	chdir := func(path string) {
		err := f.Evaler.Chdir(path)
		if err != nil {
			t.Skip("chdir:", err)
		}
	}
	chdir("bin")
	chdir("../ws1/bin")

	entries, err := f.Store.Dirs(map[string]struct{}{})
	if err != nil {
		t.Error("unable to list dir history:", err)
	}
	dirs := make([]string, len(entries))
	for i, entry := range entries {
		dirs[i] = entry.Path
	}

	wantDirs := []string{
		filepath.Join(f.Home, "bin"),
		filepath.Join(f.Home, "ws1", "bin"),
		filepath.Join("ws", "bin"),
	}

	sort.Strings(dirs)
	sort.Strings(wantDirs)
	if !reflect.DeepEqual(dirs, wantDirs) {
		t.Errorf("got dirs %v, want %v", dirs, wantDirs)
	}
}

func bbAddon(name string) *ui.BufferBuilder {
	return ui.NewBufferBuilder(30).
		WritePlain("~> ").Newline().
		WriteStyled(layout.ModeLine(name, true)).SetDotToCursor().Newline()
}
