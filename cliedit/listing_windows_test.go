package cliedit

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/ui"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/util"
)

func TestLocationAddon(t *testing.T) {
	f := setupWithOpt(setupOpt{StoreOp: func(s storedefs.Store) {
		s.AddDir(`C:\usr\bin`, 1)
		s.AddDir(`C:\tmp`, 1)
		s.AddDir(`C:\home\elf`, 1)
	}})
	f.TTYCtrl.SetSize(24, 30) // Set width to 30
	defer f.Cleanup()

	evals(f.Evaler,
		`edit:location:pinned = ['C:\opt']`,
		`edit:location:hidden = ['C:\tmp']`)
	f.TTYCtrl.Inject(term.K('L', ui.Ctrl))

	wantBuf := bbAddon("LOCATION").
		WriteMarkedLines(
			`  * C:\opt                    `, styles,
			"##############################",
			` 10 C:\home\elf`,
			` 10 C:\usr\bin`,
		).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestLocationAddon_Workspace(t *testing.T) {
	f := setupWithOpt(setupOpt{StoreOp: func(s storedefs.Store) {
		s.AddDir(`C:\usr\bin`, 1)
		s.AddDir(`ws\bin`, 1)
		s.AddDir(`other-ws\bin`, 1)
	}})
	defer f.Cleanup()
	util.ApplyDir(
		util.Dir{
			"ws1": util.Dir{
				"bin": util.Dir{},
				"tmp": util.Dir{}}})
	err := os.Chdir(`ws1\tmp`)
	if err != nil {
		t.Skip("chdir:", err)
	}
	f.TTYCtrl.SetSize(24, 30) // Set width to 30

	evals(f.Evaler,
		`edit:location:workspaces = [&ws='`+
			regexp.QuoteMeta(f.Home)+`\\'ws.]`)

	f.TTYCtrl.Inject(term.K('L', ui.Ctrl))
	wantBuf := term.NewBufferBuilder(30).
		Write(`~\ws1\tmp> `).Newline().
		WriteStyled(layout.ModeLine("LOCATION", true)).SetDotHere().Newline().
		WriteMarkedLines(
			` 10 ws\bin                    `, styles,
			"##############################",
			` 10 C:\usr\bin`,
		).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)

	f.TTYCtrl.Inject(term.K(ui.Enter))
	wantBuf = term.NewBufferBuilder(30).
		Write(`~\ws1\bin> `).SetDotHere().Buffer()
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
	evals(f.Evaler,
		`edit:location:workspaces = [&ws='`+
			regexp.QuoteMeta(f.Home)+`\\'ws.]`)

	chdir := func(path string) {
		err := f.Evaler.Chdir(path)
		if err != nil {
			t.Skip("chdir:", err)
		}
	}
	chdir("bin")
	chdir(`..\ws1\bin`)

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
