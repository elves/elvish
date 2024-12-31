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
# This script relies on the fact that go tests are always run in the package's
# directory; it should be run before changing directory.
#
# Note: Using "-o $workdir/" instead of "-o $workdir/elvish" here gets us the
# correct file extension on Windows
go build (if (not-eq $E:GOCOVERDIR '') { put -cover }) -o $workdir/ src.elv.sh/$entrypoint
`

func TestTranscripts_Default(t *testing.T) {
	testTranscripts(t, "cmd/elvish")
}

func TestTranscripts_NoDaemon(t *testing.T) {
	testTranscripts(t, "cmd/nodaemon/elvish")
}

func TestTranscripts_WithPprof(t *testing.T) {
	testTranscripts(t, "cmd/withpprof/elvish")
}

func testTranscripts(t *testing.T, entrypoint string) {
	workdir := t.TempDir()
	err := eval.NewEvaler().Eval(
		parse.Source{Name: "[build]", Code: buildScript},
		eval.EvalCfg{
			Global: eval.BuildNs().AddVars(map[string]vars.Var{
				"entrypoint": vars.NewReadOnly(entrypoint),
				"workdir":    vars.NewReadOnly(workdir),
			}).Ns()})
	if err != nil {
		t.Fatal(err)
	}
	testutil.Chdir(t, workdir)

	evaltest.TestTranscriptsInFS(t, transcripts)
}
