package styledown_test

import (
	"embed"
	"testing"

	"src.elv.sh/pkg/eval/evaltest"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	// Both Render and Derender are available as Elvish builtins, so no
	// additional setup is needed.
	evaltest.TestTranscriptsInFS(t, transcripts)
}
