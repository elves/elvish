//go:build !windows && !plan9 && !js

package unix

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
)

var (
	That             = evaltest.That
	ErrorWithMessage = evaltest.ErrorWithMessage
)

func useUNIX(ev *eval.Evaler) {
	ev.ExtendGlobal(eval.BuildNs().AddNs("unix", Ns))
}
