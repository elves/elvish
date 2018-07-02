package program

import (
	"fmt"
	"os"

	"github.com/elves/elvish/buildinfo"
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

// ShowBuildInfo shows build information.
type ShowBuildInfo struct {
	JSON bool
}

func (info ShowBuildInfo) Main([]string) int {
	if info.JSON {
		fmt.Printf(`{"version": %s, "goroot": %s, "gopath": %s}`+"\n",
			quoteJSON(buildinfo.Version), quoteJSON(buildinfo.GoRoot),
			quoteJSON(buildinfo.GoPath))
	} else {
		fmt.Println("Version:", buildinfo.Version)
		fmt.Println("GOROOT at build time:", buildinfo.GoRoot)
		fmt.Println("GOPATH at build time:", buildinfo.GoPath)
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
