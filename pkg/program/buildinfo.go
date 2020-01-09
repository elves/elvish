package program

import (
	"fmt"
	"os"
	"runtime"

	"github.com/elves/elvish/pkg/buildinfo"
)

// ShowVersion shows the version.
type ShowVersion struct{}

func (ShowVersion) Main(fds [3]*os.File, _ []string) int {
	fmt.Fprintln(fds[1], buildinfo.Version)
	return 0
}

// ShowBuildInfo shows build information.
type ShowBuildInfo struct {
	JSON bool
}

func (info ShowBuildInfo) Main(fds [3]*os.File, _ []string) int {
	if info.JSON {
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
