package runtime_test

import (
	"embed"
	"errors"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/mods/runtime"
	"src.elv.sh/pkg/testutil"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		// We can't rely on the default runtime module installed by evaltest
		// because the runtime modules reads Evaler fields during
		// initialization.
		"use-runtime-good-paths", func(t *testing.T, ev *eval.Evaler) {
			testutil.Set(t, runtime.OSExecutable,
				func() (string, error) { return "/path/to/elvish", nil })
			ev.LibDirs = []string{"/lib/1", "/lib/2"}
			ev.RcPath = "/path/to/rc.elv"
			ev.EffectiveRcPath = "/path/to/effective/rc.elv"

			ev.ExtendGlobal(eval.BuildNs().AddNs("runtime", runtime.Ns(ev)))
		},
		"use-runtime-bad-paths", func(t *testing.T, ev *eval.Evaler) {
			testutil.Set(t, runtime.OSExecutable,
				func() (string, error) { return "bad", errors.New("bad") })
			// Leaving all the path fields in ev unset

			ev.ExtendGlobal(eval.BuildNs().AddNs("runtime", runtime.Ns(ev)))
		},
	)
}
