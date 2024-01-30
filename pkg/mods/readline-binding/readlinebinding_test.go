package readline_binding_test

import (
	"embed"
	"os"
	"testing"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/edit"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/mods"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"prepare-deps",
		func(ev *eval.Evaler) {
			mods.AddTo(ev)
			ed := edit.NewEditor(cli.NewTTY(os.Stdin, os.Stderr), ev, nil)
			ev.ExtendBuiltin(eval.BuildNs().AddNs("edit", ed))
		})
}
