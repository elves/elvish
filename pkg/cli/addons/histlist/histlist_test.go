package histlist

import (
	"errors"
	"fmt"
	"testing"

	"github.com/elves/elvish/pkg/cli"
	. "github.com/elves/elvish/pkg/cli/apptest"
	"github.com/elves/elvish/pkg/cli/histutil"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/store"
	"github.com/elves/elvish/pkg/ui"
)

func TestStart_NoStore(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{})
	f.TestTTYNotes(t, "no history store")
}

type faultyStore struct{}

var mockError = errors.New("mock error")

func (s faultyStore) AllCmds() ([]store.Cmd, error) { return nil, mockError }

func TestStart_StoreError(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{Store: faultyStore{}})
	f.TestTTYNotes(t, "db error: mock error")
}

func TestStart_OK(t *testing.T) {
	f := Setup()
	defer f.Stop()

	st := histutil.NewMemoryStore()
	st.AddCmd(store.Cmd{Text: "foo", Seq: 0})
	st.AddCmd(store.Cmd{Text: "bar", Seq: 1})
	st.AddCmd(store.Cmd{Text: "baz", Seq: 2})
	Start(f.App, Config{Store: st})

	// Test UI.
	f.TTY.TestBuffer(t,
		makeListingBuf(
			" HISTORY (dedup on) ", "",
			"   0 foo",
			"   1 bar",
			"   2 baz"))

	// Test filtering.
	f.TTY.Inject(term.K('b'))
	f.TTY.TestBuffer(t,
		makeListingBuf(
			" HISTORY (dedup on) ", "b",
			"   1 bar",
			"   2 baz"))

	// Test accepting.
	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t, "baz", term.DotHere)

	// Test accepting when there is already some text.
	st.AddCmd(store.Cmd{Text: "baz2"})
	Start(f.App, Config{Store: st})
	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t, "baz",
		// codearea now contains newly inserted entry on a separate line
		"\n", "baz2", term.DotHere)
}

func TestStart_Dedup(t *testing.T) {
	f := Setup()
	defer f.Stop()

	st := histutil.NewMemoryStore()
	st.AddCmd(store.Cmd{Text: "ls", Seq: 0})
	st.AddCmd(store.Cmd{Text: "echo", Seq: 1})
	st.AddCmd(store.Cmd{Text: "ls", Seq: 2})

	// No dedup
	Start(f.App, Config{Store: st, Dedup: func() bool { return false }})
	f.TTY.TestBuffer(t,
		makeListingBuf(
			" HISTORY ", "",
			"   0 ls",
			"   1 echo",
			"   2 ls"))
	f.App.MutateState(func(s *cli.State) { s.Addon = nil })

	// With dedup
	Start(f.App, Config{Store: st, Dedup: func() bool { return true }})
	f.TTY.TestBuffer(t,
		makeListingBuf(
			" HISTORY (dedup on) ", "",
			"   1 echo",
			"   2 ls"))
}

func TestStart_CaseSensitive(t *testing.T) {
	f := Setup()
	defer f.Stop()

	st := histutil.NewMemoryStore()
	st.AddCmd(store.Cmd{Text: "ls", Seq: 0})
	st.AddCmd(store.Cmd{Text: "LS", Seq: 1})

	// Case sensitive
	Start(f.App, Config{Store: st, CaseSensitive: func() bool { return true }})
	f.TTY.Inject(term.K('l'))
	f.TTY.TestBuffer(t,
		makeListingBuf(
			" HISTORY (dedup on) ", "l",
			"   0 ls"))
	f.App.MutateState(func(s *cli.State) { s.Addon = nil })

	// Case insensitive
	Start(f.App, Config{Store: st, CaseSensitive: func() bool { return false }})
	f.TTY.Inject(term.K('l'))
	f.TTY.TestBuffer(t,
		makeListingBuf(
			" HISTORY (dedup on) (case-insensitive) ", "l",
			"   0 ls",
			"   1 LS"))
}

func bb() *term.BufferBuilder { return term.NewBufferBuilder(50) }

func makeListingBuf(mode, filter string, lines ...string) *term.Buffer {
	b := bb().Newline().
		WriteStyled(cli.ModeLine(mode, true)).
		Write(filter).SetDotHere()
	for i, line := range lines {
		b.Newline()
		if i < len(lines)-1 {
			b.Write(line)
		} else {
			b.WriteStyled(
				ui.T(fmt.Sprintf("%-50s", line), ui.Inverse))
		}
	}
	return b.Buffer()
}
