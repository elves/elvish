package store

import (
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/store"
	"src.elv.sh/pkg/store/storedefs"
	"src.elv.sh/pkg/testutil"
)

func TestStore(t *testing.T) {
	testutil.InTempDir(t)
	s, err := store.NewStore("db")
	if err != nil {
		t.Fatal(err)
	}
	ns := Ns(s)

	setup := func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddNs("store", ns))
	}
	TestWithEvalerSetup(t, setup,
		// Add commands
		That("store:next-cmd-seq").Puts(1),
		That("store:add-cmd foo").Puts(1),
		That("store:add-cmd bar").Puts(2),
		That("store:add-cmd baz").Puts(3),
		That("store:next-cmd-seq").Puts(4),
		// Query commands
		That("store:cmd 1").Puts("foo"),
		That("store:cmds 1 4").Puts(cmd("foo", 1), cmd("bar", 2), cmd("baz", 3)),
		That("store:cmds 2 3").Puts(cmd("bar", 2)),
		That("store:next-cmd 1 f").Puts(cmd("foo", 1)),
		That("store:prev-cmd 3 b").Puts(cmd("bar", 2)),
		// Delete commands
		That("store:del-cmd 2").DoesNothing(),
		That("store:cmds 1 4").Puts(cmd("foo", 1), cmd("baz", 3)),

		// Add directories
		That("store:add-dir /foo").DoesNothing(),
		That("store:add-dir /bar").DoesNothing(),
		// Query directories
		That("store:dirs").Puts(
			dir("/bar", store.DirScoreIncrement),
			dir("/foo", store.DirScoreIncrement*store.DirScoreDecay)),
		// Delete directories
		That("store:del-dir /foo").DoesNothing(),
		That("store:dirs").Puts(
			dir("/bar", store.DirScoreIncrement)),
	)
}

func cmd(s string, i int) storedefs.Cmd     { return storedefs.Cmd{Text: s, Seq: i} }
func dir(s string, f float64) storedefs.Dir { return storedefs.Dir{Path: s, Score: f} }
