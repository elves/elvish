package buildinfo

import (
	"fmt"
	"testing"

	. "src.elv.sh/pkg/prog/progtest"
)

func TestProgram(t *testing.T) {
	Test(t, &Program{},
		ThatElvish("-version").WritesStdout(Value.Version+"\n"),
		ThatElvish("-version", "-json").WritesStdout(mustToJSON(Value.Version)+"\n"),

		ThatElvish("-buildinfo").WritesStdout(
			fmt.Sprintf(
				"Version: %v\nGo version: %v\nReproducible build: %v\n",
				Value.Version, Value.GoVersion, Value.Reproducible)),
		ThatElvish("-buildinfo", "-json").WritesStdout(mustToJSON(Value)+"\n"),

		ThatElvish().ExitsWith(2).WritesStderr("internal error: no suitable subprogram\n"),
	)
}
