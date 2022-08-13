package runtime

import (
	"os"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
)

var That = evaltest.That

func TestRuntime(t *testing.T) {
	setup := func(ev *eval.Evaler) {
		ev.LibDirs = []string{"/lib/1", "/lib/2"}
		ev.RcPath = "/path/to/rc.elv"
		ev.EffectiveRcPath = "/path/to/effective/rc.elv"

		ev.ExtendGlobal(eval.BuildNs().AddNs("runtime", Ns(ev)))
	}

	elvishPath, _ := os.Executable()

	evaltest.TestWithSetup(t, setup,
		That("put $runtime:lib-dirs").Puts(vals.MakeList("/lib/1", "/lib/2")),
		That("put $runtime:rc-path").Puts("/path/to/rc.elv"),
		That("put $runtime:effective-rc-path").Puts("/path/to/effective/rc.elv"),
		That(`put $runtime:elvish-path`).Puts(elvishPath),
	)
}

func TestRuntime_NilPath(t *testing.T) {
	setup := func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddNs("runtime", Ns(ev)))
	}
	evaltest.TestWithSetup(t, setup,
		That("put $runtime:rc-path").Puts(nil),
		That("put $runtime:effective-rc-path").Puts(nil),
	)
}
