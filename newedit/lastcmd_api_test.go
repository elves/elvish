package newedit

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/newedit/listing"
	"github.com/elves/elvish/newedit/types"
)

func TestInitLastCmd_Start(t *testing.T) {
	ed := &fakeEditor{}
	ev := eval.NewEvaler()
	lsMode := listing.Mode{}

	ns := initLastcmd(ed, ev, &lsMode)

	// Call <edit:listing>:start.
	fm := eval.NewTopFrame(ev, eval.NewInternalSource("[test]"), nil)
	fm.Call(getFn(ns, "start"), nil, eval.NoOpts)

	// Verify that the current mode supports listing.
	lister, ok := ed.state.Mode().(types.Lister)
	if !ok {
		t.Errorf("Mode is not Lister after <edit:listing>:start")
	}
	// Verify the listing.
	buf := ui.Render(lister.List(10), 20)
	wantBuf := ui.NewBufferBuilder(20).
		WriteString("    echo hello world", "7").Newline().
		WriteUnstyled("  0 echo").Newline().
		WriteUnstyled("  1 hello").Newline().
		WriteUnstyled("  2 world").Buffer()
	if !reflect.DeepEqual(buf, wantBuf) {
		t.Errorf("Rendered listing is %v, want %v", buf, wantBuf)
	}
}
