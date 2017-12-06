package program

import (
	"fmt"
	"os"

	"github.com/elves/elvish/build"
)

// ShowHelp shows help message.
type ShowHelp struct {
	// Whether help is being shown because user invoked Elvish in a wrong way
	// (i.e. with bad flags or arguments).
	WrongUsage bool
}

func (ShowHelp) Main([]string) int {
	usage()
	return 0
}

type ShowCorrectUsage struct{}

func (ShowCorrectUsage) Main([]string) int {
	usage()
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
		fmt.Printf("{version: %s, builder: %s}\n",
			quoteJSON(build.Version), quoteJSON(build.Builder))
	} else {
		fmt.Println("version:", build.Version)
		fmt.Println("builder:", build.Builder)
	}
	return 0
}
