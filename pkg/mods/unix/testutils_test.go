//go:build unix

package unix

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
)

var (
	That             = evaltest.That
	ErrorWithMessage = evaltest.ErrorWithMessage
)

func useUnix(ev *eval.Evaler) {
	ev.ExtendGlobal(eval.BuildNs().AddNs("unix", Ns))
}
