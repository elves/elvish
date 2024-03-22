package elvdoc_test

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/elvdoc"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/ui"
)

var Dedent = testutil.Dedent
var stylesheet = ui.RuneStylesheet{
	'v': ui.FgGreen, '$': ui.FgMagenta,
}

var highlightCodeBlockTests = []struct {
	name string
	info string
	code string
	want ui.Text
}{
	{
		name: "elvish",
		info: "elvish extra info",
		code: "echo $pwd",
		want: ui.MarkLines(
			"echo $pwd", stylesheet,
			"vvvv $$$$"),
	},
	{
		name: "elvish-transcript",
		info: "elvish-transcript extra info",
		code: Dedent(`
			~> echo $pwd
			/home/elf
			`),
		want: ui.MarkLines(
			"~> echo $pwd\n", stylesheet,
			"   vvvv $$$$",
			"/home/elf\n",
		),
	},
	{
		name: "elvish-transcript multi line",
		info: "elvish-transcript extra info",
		code: Dedent(`
			~> echo $pwd
			   echo $pwd
			/home/elf
			/home/elf
			`),
		want: ui.MarkLines(
			"~> echo $pwd\n", stylesheet,
			"   vvvv $$$$",
			"   echo $pwd\n", stylesheet,
			"   vvvv $$$$",
			"/home/elf\n",
			"/home/elf\n",
		),
	},
	{
		name: "elvish-transcript suppress comment/directive",
		info: "elvish-transcript extra info",
		code: Dedent(`
			//dir
			// A comment
			~> echo $pwd
			/home/elf
			`),
		want: ui.MarkLines(
			"~> echo $pwd\n", stylesheet,
			"   vvvv $$$$",
			"/home/elf\n",
		),
	},
	{
		name: "other languages",
		info: "bash",
		code: "echo $pwd",
		want: ui.T("echo $pwd"),
	},
}

func TestHighlightCodeBlock(t *testing.T) {
	for _, tc := range highlightCodeBlockTests {
		t.Run(tc.name, func(t *testing.T) {
			got := elvdoc.HighlightCodeBlock(tc.info, tc.code)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}
