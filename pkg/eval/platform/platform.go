// Package `os` exposes variables and functions that deal with the specific
// platform being run on; such as the OS name and CPU architecture.
package platform

import (
	"runtime"

	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vars"
)

//elvdoc:var arch
//
// The architecture of the platform; e.g., amd64, arm, ppc.
// This corresponds to Go's
// [`GOARCH`](https://pkg.go.dev/runtime?tab=doc#pkg-constants) constant.
// This is read-only.

//elvdoc:var os
//
// The name of the operating system; e.g., darwin (macOS), linux, etc.
// This corresponds to Go's
// [`GOOS`](https://pkg.go.dev/runtime?tab=doc#pkg-constants) constant.
// This is read-only.

var Ns = eval.Ns{
	"arch": vars.NewReadOnly(runtime.GOARCH),
	"os": vars.NewReadOnly(runtime.GOOS),
}
