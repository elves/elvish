//go:build unix

package unix

import (
	"errors"
	"testing"

	"golang.org/x/sys/unix"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/testutil"
)

func TestRlimits(t *testing.T) {
	testutil.Set(t, &getRlimit, func(res int, lim *unix.Rlimit) error {
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
	testutil.Set(t, &setRlimit, func(res int, lim *unix.Rlimit) error {
		switch res {
		case unix.RLIMIT_CPU:
			cpuCur = rlimTToInt(lim.Cur)
			cpuMax = rlimTToInt(lim.Max)
		case unix.RLIMIT_NOFILE:
			return errors.New("fake setrlimit error")
		}
		return nil
	})

	setup := func(ev *eval.Evaler) {
		useUnix(ev)
		ev.ExtendGlobal(eval.BuildNs().
			AddVar("cpu-cur", vars.FromPtr(&cpuCur)).
			AddVar("cpu-max", vars.FromPtr(&cpuMax)))
	}

	evaltest.TestWithEvalerSetup(t, setup,
		That("put $unix:rlimits[cpu]").Puts(vals.EmptyMap),
		That("put $unix:rlimits[nofile]").Puts(vals.MakeMap("cur", 30, "max", 40)),
		That("has-key $unix:rlimits stack").Puts(false),

		That("set unix:rlimits[cpu] = [&cur=3 &max=8]", "put $cpu-cur $cpu-max").
			Puts(3, 8),
		That("set unix:rlimits[cpu] = [&cur=4]", "put $cpu-cur $cpu-max").
			Puts(4, -1),
		That("set unix:rlimits[cpu] = [&]", "put $cpu-cur $cpu-max").Puts(-1, -1),

		That("set unix:rlimits[nofile] = [&]").
			Throws(ErrorWithMessage("setrlimit nofile: fake setrlimit error")),

		// Error parsing the rlimits map
		That("set unix:rlimits = x").
			Throws(errs.BadValue{What: "$unix:rlimits", Valid: "map", Actual: "string"}),
		That("set unix:rlimits = [&[]=[&]]").
			Throws(errs.BadValue{What: "key of $unix:rlimits",
				Valid: "string", Actual: "list"}),
		That("set unix:rlimits = [&bad-resource=[&]]").
			Throws(errs.BadValue{What: "key of $unix:rlimits",
				Valid: "valid resource key", Actual: "bad-resource"}),
		That("set unix:rlimits = [&]").
			Throws(errs.BadValue{What: "$unix:rlimits",
				Valid: "map containing all resource keys", Actual: "[&]"}),
		// Error parsing a value of the rlimits map
		That("set unix:rlimits[cpu] = x").
			Throws(errs.BadValue{What: "rlimit value", Valid: "map", Actual: "string"}),
		That("set unix:rlimits[cpu] = [&bad]").
			Throws(errs.BadValue{What: "key of rlimit value",
				Valid: "cur or max", Actual: "bad"}),
		That("set unix:rlimits[cpu] = [&cur=[]]").
			Throws(errs.BadValue{What: "cur in rlimit value",
				Valid: rlimTValid, Actual: "[]"}),
		That("set unix:rlimits[cpu] = [&cur=1 &max=[]]").
			Throws(errs.BadValue{What: "max in rlimit value",
				Valid: rlimTValid, Actual: "[]"}),
	)
}

func rlimTToInt(r rlimT) int {
	if r == unix.RLIM_INFINITY {
		return -1
	}
	return int(r)
}
