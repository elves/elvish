package doc_test

import (
	"embed"
	"io/fs"
	"testing"

	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/mods/doc"
)

//go:embed fakepkg
var fakepkg embed.FS

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	// The result of reading the FS is cached. As a result, this override can't
	// be reverted, so we just do it here instead of properly inside a setup
	// function.
	*doc.ElvFiles, _ = fs.Sub(fakepkg, "fakepkg")
	evaltest.TestTranscriptsInFS(t, transcripts)
}
