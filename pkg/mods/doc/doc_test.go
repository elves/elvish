package doc_test

import (
	"embed"
	"io/fs"
	"testing"

	"src.elv.sh/pkg/elvdoc"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/md"
	"src.elv.sh/pkg/mods/doc"
	"src.elv.sh/pkg/testutil"
)

var (
	That             = evaltest.That
	ErrorWithMessage = evaltest.ErrorWithMessage

	Dedent = testutil.Dedent
)

var (
	//go:embed fakepkg
	fakepkg embed.FS
	//go:embed fakepkg/mods/foo/variable.md
	fooVariableDoc string
	//go:embed fakepkg/mods/foo/function.md
	fooFunctionDoc string
	//go:embed fakepkg/eval/break.md
	breakDoc string
	//go:embed fakepkg/eval/num.md
	numDoc string
)

func init() {
	*doc.ElvFiles, _ = fs.Sub(fakepkg, "fakepkg")
}

func TestShow(t *testing.T) {
	evaltest.TestWithEvalerSetup(t, setupDoc,
		That("doc:show foo:function").Prints(render(fooFunctionDoc, 80)),
		That("doc:show &width=30 foo:function").Prints(render(fooFunctionDoc, 30)),
		// TODO: Test in pty
		That("doc:show '$foo:variable'").Prints(render(fooVariableDoc, 80)),

		That("doc:show break").Prints(render(breakDoc, 80)),
		That("doc:show builtin:break").Prints(render(breakDoc, 80)),
		// Test that relative links to language.html are converted to absolute
		// links to https://elv.sh/ref/language.html, but other relative links
		// are not.
		That("doc:show num").Prints(render(numDoc, 80)),

		That("doc:show foo:bad").Throws(ErrorWithMessage("no doc for foo:bad")),
		That("doc:show bad:foo").Throws(ErrorWithMessage("no doc for bad:foo")),
	)
}

func TestFind(t *testing.T) {
	evaltest.TestWithEvalerSetup(t, setupDoc,
		That("doc:find ipsum").Prints(highlightBraced(Dedent(`
			foo:function:
			  … Lorem {ipsum} dolor sit amet. …
			$foo:variable:
			  … Lorem {ipsum}.
			`))),
	)
}

func TestSource(t *testing.T) {
	evaltest.TestWithEvalerSetup(t, setupDoc,
		That("doc:source '$foo:variable'").Puts(fooVariableDoc),
		// The implementation of doc:source is used by doc:show internally, so
		// we only test a simple case here.
	)
}

func TestSymbols(t *testing.T) {
	evaltest.TestWithEvalerSetup(t, setupDoc,
		That("doc:-symbols").Puts(
			// All symbols, sorted
			"$foo:variable", "break", "foo:function", "num"),

		That("doc:-symbols >&-").Throws(eval.ErrPortDoesNotSupportValueOutput),
	)
}

func setupDoc(ev *eval.Evaler) {
	ev.ExtendGlobal(eval.BuildNs().AddNs("doc", doc.Ns))
}

func render(s string, w int) string {
	return md.RenderString(s, &md.TTYCodec{Width: w, HighlightCodeBlock: elvdoc.HighlightCodeBlock})
}
