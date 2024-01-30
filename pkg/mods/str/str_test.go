package str

import (
	"embed"
	"testing"

	"src.elv.sh/pkg/eval/evaltest"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"use-str", evaltest.Use("str", Ns),
	)
}
