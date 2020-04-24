package prog

import (
	"encoding/json"
	"fmt"
	"runtime"
	"testing"

	"github.com/elves/elvish/pkg/buildinfo"
	. "github.com/elves/elvish/pkg/prog/progtest"
)

func TestVersion(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	Main(f.Fds(), elvish("-version"))

	f.TestOut(t, 1, buildinfo.Version+"\n")
}

func TestBuildInfo(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	Main(f.Fds(), elvish("-buildinfo"))

	f.TestOut(t, 1,
		fmt.Sprintf(
			"Version: %v\nGo version: %v\nReproducible build: %v\n",
			buildinfo.Version,
			runtime.Version(),
			buildinfo.Reproducible))
}

func TestBuildInfo_JSON(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	Main(f.Fds(), elvish("-buildinfo", "-json"))

	f.TestOut(t, 1,
		mustToJSON(struct {
			Version      string `json:"version"`
			GoVersion    string `json:"goversion"`
			Reproducible bool   `json:"reproducible"`
		}{
			buildinfo.Version,
			runtime.Version(),
			buildinfo.Reproducible == "true",
		})+"\n")
}

func TestHelp(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	Main(f.Fds(), elvish("-help"))

	f.TestOutSnippet(t, 1, "Usage: elvish [flags] [script]")
}

func TestBadFlag(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := Main(f.Fds(), elvish("-bad-flag"))

	testError(t, f, exit, "flag provided but not defined: -bad-flag")
}

func TestWeb_SpuriousArgument(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := Main(f.Fds(), elvish("-web", "x"))

	testError(t, f, exit, "arguments are not allowed with -web")
}

func TestWeb_SpuriousC(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := Main(f.Fds(), elvish("-web", "-c"))

	testError(t, f, exit, "-c cannot be used together with -web")
}

func TestDaemon_SpuriousArgument(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := Main(f.Fds(), elvish("-daemon", "x"))

	testError(t, f, exit, "arguments are not allowed with -daemon")
}

func elvish(args ...string) []string {
	return append([]string{"elvish"}, args...)
}

func mustToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func testError(t *testing.T, f *Fixture, exit int, wantErrSnippet string) {
	t.Helper()
	if exit != 2 {
		t.Errorf("got exit %v, want 2", exit)
	}
	f.TestOutSnippet(t, 2, wantErrSnippet)
}
