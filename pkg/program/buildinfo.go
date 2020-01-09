package program

import (
	"fmt"
	"os"
	"runtime"

	"github.com/elves/elvish/pkg/buildinfo"
)

type buildInfoProgram struct{ JSON bool }

func (p buildInfoProgram) Main(fds [3]*os.File, _ []string) int {
	if p.JSON {
		fmt.Fprintf(fds[1],
			`{"version":%s,"goversion":%s,"reproducible":%v}`+"\n",
			quoteJSON(buildinfo.Version), quoteJSON(runtime.Version()),
			buildinfo.Reproducible)
	} else {
		fmt.Fprintln(fds[1], "Version:", buildinfo.Version)
		fmt.Fprintln(fds[1], "Go version:", runtime.Version())
		fmt.Fprintln(fds[1], "Reproducible build:", buildinfo.Reproducible)
	}
	return 0
}

type versionProgram struct{}

func (versionProgram) Main(fds [3]*os.File, _ []string) int {
	fmt.Fprintln(fds[1], buildinfo.Version)
	return 0
}
