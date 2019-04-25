package newedit

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/ui"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/newedit/clitypes"
	"github.com/elves/elvish/newedit/listing"
	"github.com/elves/elvish/store/storedefs"
)

func TestLocation_Start(t *testing.T) {
	ed := &fakeApp{}
	ev := eval.NewEvaler()
	lsMode := listing.Mode{}
	lsBinding := emptyBindingMap
	getDirs := func() ([]storedefs.Dir, error) {
		return []storedefs.Dir{
			{Path: "/usr/bin", Score: 20},
			{Path: "/home/elf", Score: 10},
		}, nil
	}
	cd := func(string) error { return nil }

	ns := initLocation(ed, ev, getDirs, cd, &lsMode, &lsBinding)

	// Call <edit:location>:start.
	fm := eval.NewTopFrame(ev, eval.NewInternalSource("[test]"), nil)
	fm.Call(getFn(ns, "start"), eval.NoArgs, eval.NoOpts)

	// Verify that the current mode supports listing.
	lister, ok := ed.state.Mode().(clitypes.Lister)
	if !ok {
		t.Errorf("Mode is not Lister after <edit:location>:start")
	}
	// Verify the actual listing.
	buf := ui.Render(lister.List(10), 30)
	wantBuf := ui.NewBufferBuilder(30).
		WriteString(" 20 /usr/bin", "7").Newline().
		WriteString(" 10 /home/elf", "").Buffer()
	if !reflect.DeepEqual(buf, wantBuf) {
		t.Errorf("Rendered listing is %v, want %v", buf, wantBuf)
	}
}
