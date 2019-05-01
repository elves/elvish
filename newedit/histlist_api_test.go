package newedit

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/ui"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cli/listing"
	"github.com/elves/elvish/eval"
)

func TestHistlist_Start(t *testing.T) {
	ed := &fakeApp{}
	ev := eval.NewEvaler()
	lsMode := listing.Mode{}
	lsBinding := emptyBindingMap
	// TODO: Move this into common setup.
	histFuser, err := histutil.NewFuser(testStore)
	if err != nil {
		panic(err)
	}

	ns := initHistlist(ed, ev, histFuser.AllCmds, &lsMode, &lsBinding)

	// Call <edit:histlist>:start.
	fm := eval.NewTopFrame(ev, eval.NewInternalSource("[test]"), nil)
	fm.Call(getFn(ns, "start"), eval.NoArgs, eval.NoOpts)

	// Verify that the current mode supports listing.
	lister, ok := ed.state.Mode().(clitypes.Lister)
	if !ok {
		t.Errorf("Mode is not Lister after <edit:histlist>:start")
	}
	// Verify the actual listing.
	buf := ui.Render(lister.List(10), 30)
	wantBuf := ui.NewBufferBuilder(30).
		WriteString("   0 echo hello world", "7").Buffer()
	if !reflect.DeepEqual(buf, wantBuf) {
		t.Errorf("Rendered listing is %v, want %v", buf, wantBuf)
	}
}
