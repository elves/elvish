package e2e_test

import (
	"embed"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/testutil"
)

//go:embed *.elvts
var transcripts embed.FS

const buildScript = `
# Using "-o $workdir/" instead of "-o $workdir/elvish" gets us the correct file
# extension on Windows
go build (if (not-eq $E:GOCOVERDIR '') { put -cover }) -o $workdir/ src.elv.sh/cmd/elvish
`

func TestTranscripts(t *testing.T) {
	workdir := t.TempDir()
	err := eval.NewEvaler().Eval(
		parse.Source{Name: "[build]", Code: buildScript},
		eval.EvalCfg{
			Global: eval.BuildNs().AddVars(map[string]vars.Var{
				"workdir": vars.NewReadOnly(workdir),
			}).Ns()})
	if err != nil {
		t.Fatal(err)
	}
	testutil.Chdir(t, workdir)

	evaltest.TestTranscriptsInFS(t, transcripts)
}
