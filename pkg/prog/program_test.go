package prog

import (
	"encoding/json"
	"testing"

	. "github.com/elves/elvish/pkg/prog/progtest"
)

func TestHelp(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	Run(f.Fds(), Elvish("-help"))

	f.TestOutSnippet(t, 1, "Usage: elvish [flags] [script]")
}

func TestBadFlag(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := Run(f.Fds(), Elvish("-bad-flag"))

	TestError(t, f, exit, "flag provided but not defined: -bad-flag")
}

func mustToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
