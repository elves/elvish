package buildinfo

import (
	"fmt"
	"testing"

	"src.elv.sh/pkg/prog"
	. "src.elv.sh/pkg/prog/progtest"
)

func TestVersion(t *testing.T) {
	f := Setup(t)

	prog.Run(f.Fds(), Elvish("-version"), Program)

	f.TestOut(t, 1, Value.Version+"\n")
	f.TestOut(t, 2, "")
}

func TestVersion_JSON(t *testing.T) {
	f := Setup(t)

	prog.Run(f.Fds(), Elvish("-version", "-json"), Program)

	f.TestOut(t, 1, mustToJSON(Value.Version)+"\n")
	f.TestOut(t, 2, "")
}

func TestBuildInfo(t *testing.T) {
	f := Setup(t)

	prog.Run(f.Fds(), Elvish("-buildinfo"), Program)

	f.TestOut(t, 1,
		fmt.Sprintf(
			"Version: %v\nGo version: %v\nReproducible build: %v\n",
			Value.Version, Value.GoVersion, Value.Reproducible))
	f.TestOut(t, 2, "")
}

func TestBuildInfo_JSON(t *testing.T) {
	f := Setup(t)

	prog.Run(f.Fds(), Elvish("-buildinfo", "-json"), Program)

	f.TestOut(t, 1, mustToJSON(Value)+"\n")
	f.TestOut(t, 2, "")
}
