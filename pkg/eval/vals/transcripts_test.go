package vals_test

import (
	"embed"
	"fmt"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"scan-string-slice-in-global", evaltest.GoFnInGlobal("scan-string-slice",
			func(fm *eval.Frame, s []string) {
				fmt.Fprintf(fm.ByteOutput(), "scanned: %#v\n", s)
			}),
	)
}
