package buildinfo

import (
	"fmt"
	"runtime"
	"testing"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/prog/progtest"
)

var ThatElvish = progtest.ThatElvish

func TestProgram(t *testing.T) {
	progtest.Test(t, &Program{},
		ThatElvish("-version").WritesStdout(Value.Version+"\n"),
		ThatElvish("-version", "-json").WritesStdout(mustToJSON(Value.Version)+"\n"),

		ThatElvish("-buildinfo").WritesStdout(
			fmt.Sprintf(
				"Version: %v\nGo version: %v\n", Value.Version, Value.GoVersion)),
		ThatElvish("-buildinfo", "-json").WritesStdout(mustToJSON(Value)+"\n"),

		ThatElvish().ExitsWith(2).WritesStderr("internal error: no suitable subprogram\n"),
	)
}

func TestAddVariant(t *testing.T) {
	got := addVariant("0.42.0", "")
	want := "0.42.0"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got = addVariant("0.42.0", "distro")
	want = "0.42.0+distro"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestValue(t *testing.T) {
	vals.TestValue(t, Value).
		Index("version", Value.Version).
		Index("go-version", runtime.Version())
}
