package path_test

import (
	"embed"
	"testing"

	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"in-temp-dir-with-d-f", func(t *testing.T) {
			testutil.InTempDir(t)
			testutil.ApplyDir(testutil.Dir{
				"d": testutil.Dir{
					"f": "",
				},
			})
		},
	)
}
