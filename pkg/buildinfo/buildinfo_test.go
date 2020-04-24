package buildinfo

import (
	"encoding/json"
	"fmt"
	"runtime"
	"testing"

	"github.com/elves/elvish/pkg/prog"
	. "github.com/elves/elvish/pkg/prog/progtest"
)

func TestVersion(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	prog.Run(f.Fds(), Elvish("-version"), Program)

	f.TestOut(t, 1, Version+"\n")
}

func TestBuildInfo(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	prog.Run(f.Fds(), Elvish("-buildinfo"), Program)

	f.TestOut(t, 1,
		fmt.Sprintf(
			"Version: %v\nGo version: %v\nReproducible build: %v\n",
			Version,
			runtime.Version(),
			Reproducible))
}

func TestBuildInfo_JSON(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	prog.Run(f.Fds(), Elvish("-buildinfo", "-json"), Program)

	f.TestOut(t, 1,
		mustToJSON(struct {
			Version      string `json:"version"`
			GoVersion    string `json:"goversion"`
			Reproducible bool   `json:"reproducible"`
		}{
			Version,
			runtime.Version(),
			Reproducible == "true",
		})+"\n")
}

func mustToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
