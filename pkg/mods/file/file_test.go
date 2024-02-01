package file_test

import (
	"embed"
	"os"
	"testing"

	"src.elv.sh/pkg/eval/evaltest"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"skip-unless-can-open", func(t *testing.T, name string) {
			if !canOpen(name) {
				t.SkipNow()
			}
		},
	)
}

func canOpen(name string) bool {
	f, err := os.Open(name)
	f.Close()
	return err == nil
}
