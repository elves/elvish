package platform_test

import (
	"embed"
	"errors"
	"testing"

	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/mods/platform"
	"src.elv.sh/pkg/testutil"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"mock-hostname", func(t *testing.T, hostname string) {
			testutil.Set(t, platform.OSHostname, func() (string, error) { return hostname, nil })
		},
		"mock-hostname-error", func(t *testing.T, msg string) {
			err := errors.New(msg)
			testutil.Set(t, platform.OSHostname, func() (string, error) { return "", err })
		},
	)
}
