package daemon_test

import (
	"embed"
	"testing"

	"src.elv.sh/pkg/daemon"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/prog/progtest"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"elvish-in-global", progtest.ElvishInGlobal(&daemon.Program{}),
	)
}
