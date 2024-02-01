//go:build unix

package unix_test

import (
	"embed"
	"errors"
	"testing"

	"golang.org/x/sys/unix"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vars"
	unixmod "src.elv.sh/pkg/mods/unix"
	"src.elv.sh/pkg/testutil"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	// Intention is to restore umask after test finishes
	testutil.Umask(t, 0)
	evaltest.TestTranscriptsInFS(t, transcripts,
		"mock-rlimit", mockRlimit,
	)
}

func mockRlimit(t *testing.T, ev *eval.Evaler) {
	testutil.Set(t, unixmod.GetRlimit, func(res int, lim *unix.Rlimit) error {
		switch res {
		case unix.RLIMIT_CPU:
			*lim = unix.Rlimit{Cur: unix.RLIM_INFINITY, Max: unix.RLIM_INFINITY}
		case unix.RLIMIT_NOFILE:
			*lim = unix.Rlimit{Cur: 30, Max: 40}
		case unix.RLIMIT_STACK:
			return errors.New("fake getrlimit error")
		}
		return nil
	})

	var cpuCur, cpuMax int
	testutil.Set(t, unixmod.SetRlimit, func(res int, lim *unix.Rlimit) error {
		switch res {
		case unix.RLIMIT_CPU:
			cpuCur = rlimTToInt(lim.Cur)
			cpuMax = rlimTToInt(lim.Max)
		case unix.RLIMIT_NOFILE:
			return errors.New("fake setrlimit error")
		}
		return nil
	})

	ev.ExtendGlobal(eval.BuildNs().
		AddVar("cpu-cur", vars.FromPtr(&cpuCur)).
		AddVar("cpu-max", vars.FromPtr(&cpuMax)))
}

func rlimTToInt(r unixmod.RlimT) int {
	if r == unix.RLIM_INFINITY {
		return -1
	}
	return int(r)
}
