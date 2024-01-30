package store

import (
	"embed"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/store"
	"src.elv.sh/pkg/testutil"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"use-store-brand-new", func(t *testing.T, ev *eval.Evaler) {
			testutil.InTempDir(t)
			s := must.OK1(store.NewStore("db"))
			ev.ExtendGlobal(eval.BuildNs().AddNs("store", Ns(s)))
		},
	)
}
