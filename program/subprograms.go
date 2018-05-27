package program

import (
	"fmt"
	"os"

	"github.com/elves/elvish/build"
	daemonsvc "github.com/elves/elvish/daemon"
	"github.com/elves/elvish/program/daemon"
)

// ShowHelp shows help message.
type ShowHelp struct {
	flag *flagSet
}

func (s ShowHelp) Main([]string) int {
	usage(os.Stdout, s.flag)
	return 0
}

type ShowCorrectUsage struct {
	message string
	flag    *flagSet
}

func (s ShowCorrectUsage) Main([]string) int {
	usage(os.Stderr, s.flag)
	return 2
}

// ShowVersion shows the version.
type ShowVersion struct{}

func (ShowVersion) Main([]string) int {
	fmt.Println(build.Version)
	fmt.Fprintln(os.Stderr, "-version is deprecated and will be removed in 0.12. Use -buildinfo instead.")
	return 0
}

// ShowBuildInfo shows build information.
type ShowBuildInfo struct {
	JSON bool
}

func (info ShowBuildInfo) Main([]string) int {
	if info.JSON {
		fmt.Printf("{\"version\": %s, \"builder\": %s}\n",
			quoteJSON(build.Version), quoteJSON(build.Builder))
	} else {
		fmt.Println("version:", build.Version)
		fmt.Println("builder:", build.Builder)
	}
	return 0
}

// Daemon runs the daemon subprogram.
type Daemon struct {
	inner *daemon.Daemon
}

func (d Daemon) Main([]string) int {
	err := d.inner.Main(daemonsvc.Serve)
	if err != nil {
		logger.Println("daemon error:", err)
		return 2
	}
	return 0
}
