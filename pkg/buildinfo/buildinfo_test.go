package buildinfo

import (
	"fmt"
	"runtime/debug"
	"testing"

	. "src.elv.sh/pkg/prog/progtest"
)

func TestProgram(t *testing.T) {
	Test(t, &Program{},
		ThatElvish("-version").WritesStdout(Value.Version+"\n"),
		ThatElvish("-version", "-json").WritesStdout(mustToJSON(Value.Version)+"\n"),

		ThatElvish("-buildinfo").WritesStdout(
			fmt.Sprintf(
				"Version: %v\nGo version: %v\n", Value.Version, Value.GoVersion)),
		ThatElvish("-buildinfo", "-json").WritesStdout(mustToJSON(Value)+"\n"),

		ThatElvish().ExitsWith(2).WritesStderr("internal error: no suitable subprogram\n"),
	)
}

var devVersionTests = []struct {
	name        string
	vcsOverride string
	bi          *debug.BuildInfo
	want        string
}{
	// next is always "0.42.0"
	{
		"no BuildInfo",
		"",
		nil,
		"0.42.0-dev.unknown",
	},
	{
		"BuildInfo with Main.Version = (devel)",
		"",
		&debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
		"0.42.0-dev.unknown",
	},
	{
		"BuildInfo with non-empty Main.Version != (devel)",
		"",
		&debug.BuildInfo{Main: debug.Module{Version: "v0.42.0-dev.foobar"}},
		"0.42.0-dev.foobar",
	},
	{
		"BuildInfo with VCS data from clean checkout",
		"",
		&debug.BuildInfo{Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890123456"},
			{Key: "vcs.time", Value: "2022-04-01T23:59:58Z"},
			{Key: "vcs.modified", Value: "false"},
		}},
		"0.42.0-dev.0.20220401235958-123456789012",
	},
	{
		"BuildInfo with VCS data from dirty checkout",
		"",
		&debug.BuildInfo{Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890123456"},
			{Key: "vcs.time", Value: "2022-04-01T23:59:58Z"},
			{Key: "vcs.modified", Value: "true"},
		}},
		"0.42.0-dev.0.20220401235958-123456789012-dirty",
	},
	{
		"BuildInfo with unknown VCS timestamp format",
		"",
		&debug.BuildInfo{Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890123456"},
			{Key: "vcs.time", Value: "April First"},
			{Key: "vcs.modified", Value: "false"},
		}},
		"0.42.0-dev.unknown",
	},
	{
		"vcsOverride",
		"20220401235958-123456789012",
		nil,
		"0.42.0-dev.0.20220401235958-123456789012",
	},
}

func TestDevVersion(t *testing.T) {
	for _, test := range devVersionTests {
		t.Run(test.name, func(t *testing.T) {
			f := func() (*debug.BuildInfo, bool) {
				if test.bi == nil {
					return nil, false
				}
				return test.bi, true
			}
			got := devVersion("0.42.0", test.vcsOverride, f)
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}
