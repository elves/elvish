package prog

import (
	"fmt"
	"os"
	"runtime"

	"github.com/elves/elvish/pkg/buildinfo"
)

type buildInfoProgram struct{}

func (p buildInfoProgram) ShouldRun(f *Flags) bool { return f.BuildInfo }

func (p buildInfoProgram) Run(fds [3]*os.File, f *Flags, _ []string) error {
	if f.JSON {
		fmt.Fprintf(fds[1],
			`{"version":%s,"goversion":%s,"reproducible":%v}`+"\n",
			quoteJSON(buildinfo.Version), quoteJSON(runtime.Version()),
			buildinfo.Reproducible)
	} else {
		fmt.Fprintln(fds[1], "Version:", buildinfo.Version)
		fmt.Fprintln(fds[1], "Go version:", runtime.Version())
		fmt.Fprintln(fds[1], "Reproducible build:", buildinfo.Reproducible)
	}
	return nil
}

type versionProgram struct{}

func (versionProgram) ShouldRun(f *Flags) bool { return f.Version }

func (versionProgram) Run(fds [3]*os.File, _ *Flags, _ []string) error {
	fmt.Fprintln(fds[1], buildinfo.Version)
	return nil
}
