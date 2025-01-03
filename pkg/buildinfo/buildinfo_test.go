package buildinfo

import (
	"runtime/debug"
	"testing"

	"src.elv.sh/pkg/testutil"
)

var devVersionTests = []struct {
	name        string
	next        string
	vcsOverride string
	buildInfo   *debug.BuildInfo
	want        string
}{
	{
		name: "no BuildInfo",
		next: "0.42.0",
		want: "0.42.0-dev.unknown",
	},
	{
		name:      "BuildInfo with Main.Version = (devel)",
		next:      "0.42.0",
		buildInfo: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
		want:      "0.42.0-dev.unknown",
	},
	{
		name:      "BuildInfo with non-empty Main.Version != (devel)",
		next:      "0.42.0",
		buildInfo: &debug.BuildInfo{Main: debug.Module{Version: "v0.42.0-dev.foobar"}},
		want:      "0.42.0-dev.foobar",
	},
	{
		name: "BuildInfo with VCS data from clean checkout",
		next: "0.42.0",
		buildInfo: &debug.BuildInfo{Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890123456"},
			{Key: "vcs.time", Value: "2022-04-01T23:59:58Z"},
			{Key: "vcs.modified", Value: "false"},
		}},
		want: "0.42.0-dev.0.20220401235958-123456789012",
	},
	{
		name: "BuildInfo with VCS data from dirty checkout",
		next: "0.42.0",
		buildInfo: &debug.BuildInfo{Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890123456"},
			{Key: "vcs.time", Value: "2022-04-01T23:59:58Z"},
			{Key: "vcs.modified", Value: "true"},
		}},
		want: "0.42.0-dev.0.20220401235958-123456789012-dirty",
	},
	{
		name: "BuildInfo with unknown VCS timestamp format",
		next: "0.42.0",
		buildInfo: &debug.BuildInfo{Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890123456"},
			{Key: "vcs.time", Value: "April First"},
			{Key: "vcs.modified", Value: "false"},
		}},
		want: "0.42.0-dev.unknown",
	},
	{
		name:        "vcsOverride",
		next:        "0.42.0",
		vcsOverride: "20220401235958-123456789012",
		want:        "0.42.0-dev.0.20220401235958-123456789012",
	},
}

func TestDevVersion(t *testing.T) {
	for _, test := range devVersionTests {
		t.Run(test.name, func(t *testing.T) {
			testutil.Set(t, &readBuildInfo,
				func() (*debug.BuildInfo, bool) {
					return test.buildInfo, test.buildInfo != nil
				})
			got := devVersion(test.next, test.vcsOverride)
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
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
