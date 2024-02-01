package file_test

import (
	"embed"
	"os"
	"testing"

	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/mods/file"
	osmod "src.elv.sh/pkg/mods/os"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"use-file", evaltest.Use("file", file.Ns),
		"use-os", evaltest.Use("os", osmod.Ns),
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
