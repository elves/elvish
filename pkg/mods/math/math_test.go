package math_test

import (
	"embed"
	"testing"

	"src.elv.sh/pkg/eval/evaltest"
)

//go:embed *.elvts *.elv
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts)
}
