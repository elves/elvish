package highlight_test

import (
	"embed"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"src.elv.sh/pkg/edit/highlight"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/ui/styledown"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"highlight-in-global", func(ev *eval.Evaler, arg string) {
			fields := strings.Fields(arg)
			maxBlock, validCommands := fields[0], fields[1:]
			evaltest.GoFnInGlobal("highlight", func(fm *eval.Frame, s string) {
				hl := highlight.NewHighlighter(highlight.Config{
					HasCommand:         func(name string) bool { return slices.Contains(validCommands, name) },
					HasCommandMaxBlock: func() time.Duration { return must.OK1(time.ParseDuration(maxBlock)) },
				})
				text, tips := hl.Get(s)
				fmt.Fprint(fm.ByteOutput(), toStyledown(text))
				for i, tip := range tips {
					fmt.Fprintf(fm.ByteOutput(), "= tip %d:\n%s", i, toStyledown(tip))
				}
			})(ev)
		},
	)
}

var styleDefs = `
? fg-bright-white bg-red
G fg-green
R fg-red
M fg-magenta
Y fg-yellow
C fg-cyan
`[1:]

func toStyledown(t ui.Text) string { return must.OK1(styledown.Derender(t, styleDefs)) }
