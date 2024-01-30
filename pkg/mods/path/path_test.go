package path_test

import (
	"embed"
	"testing"

	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/mods/file"
	osmod "src.elv.sh/pkg/mods/os"
	"src.elv.sh/pkg/mods/path"
	"src.elv.sh/pkg/mods/re"
	"src.elv.sh/pkg/mods/str"
	"src.elv.sh/pkg/testutil"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"use-path", evaltest.Use("path", path.Ns),
		"use-file", evaltest.Use("file", file.Ns),
		"use-os", evaltest.Use("os", osmod.Ns),
		"use-re", evaltest.Use("re", re.Ns),
		"use-str", evaltest.Use("str", str.Ns),
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
