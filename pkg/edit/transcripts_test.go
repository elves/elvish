package edit

import (
	"embed"
	"testing"

	"src.elv.sh/pkg/edit/complete"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	fnInGlobal := func(name string, impl any) func(*eval.Evaler) {
		return func(ev *eval.Evaler) {
			ev.ExtendBuiltin(eval.BuildNs().AddGoFn(name, impl))
		}
	}
	evaltest.TestTranscriptsInFS(t, transcripts,
		"binding-map-in-global", fnInGlobal("binding-map", makeBindingMap),
		"wordify-in-global", fnInGlobal("wordify", wordify),
		"complete-getopt-in-global", fnInGlobal("complete-getopt", completeGetopt),
		"complete-filename-in-global", fnInGlobal("complete-filename",
			wrapArgGenerator(complete.GenerateFileNames)),
		"complex-candidate-in-global", fnInGlobal("complex-candidate", complexCandidate),
		"add-var-in-global", fnInGlobal("add-var", addVar),
		"add-vars-in-global", fnInGlobal("add-vars", addVars),
		"del-var-in-global", fnInGlobal("del-var", delVar),
		"del-vars-in-global", fnInGlobal("del-vars", delVars),
	)
}
