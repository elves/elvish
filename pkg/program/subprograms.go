package program

import (
	"fmt"
	"os"
	"runtime"

	"github.com/elves/elvish/pkg/buildinfo"
	daemonsvc "github.com/elves/elvish/pkg/daemon"
	"github.com/elves/elvish/pkg/program/daemon"
)

// ShowHelp shows help message.
type ShowHelp struct {
	flag *flagSet
}

func (s ShowHelp) Main(fds [3]*os.File, _ []string) int {
	usage(fds[1], s.flag)
	return 0
}

type ShowCorrectUsage struct {
	message string
	flag    *flagSet
}

func (s ShowCorrectUsage) Main(fds [3]*os.File, _ []string) int {
	usage(fds[1], s.flag)
	return 2
}

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

// Daemon runs the daemon subprogram.
type Daemon struct {
	inner *daemon.Daemon
}

func (d Daemon) Main(fds [3]*os.File, _ []string) int {
	err := d.inner.Main(daemonsvc.Serve)
	if err != nil {
		logger.Println("daemon error:", err)
		return 2
	}
	return 0
}
