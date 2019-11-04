package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/cliutil"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/store"
)

func TestInsertLastWord(t *testing.T) {
	st, cleanupStore := store.MustGetTempStore()
	defer cleanupStore()
	st.AddCmd("echo foo bar")
	ed, _, ev, cleanup := setupWithStore(st)
	defer cleanup()

	evalf(ev, "edit:insert-last-word")
	wantBuf := codearea.Buffer{Content: "bar", Dot: 3}
	if buf := cliutil.GetCodeBuffer(ed.app); buf != wantBuf {
		t.Errorf("buf = %v, want %v", buf, wantBuf)
	}
}
