package doc_test

import (
	"io"
	"strings"
	"testing"

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

func init() {
	// The map doc.ModToCode points to is read once during lazy initialization,
	// so there is no way to undo the change.
	*doc.ModToCode = map[string]io.Reader{
		"foo:": strings.NewReader(fooModuleCode),
		"":     strings.NewReader(builtinModuleCode),
	}
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

var tildeToBackquote = strings.NewReplacer("~", "`").Replace

var (
	fooModuleCode = Dedent(`
		# A variable. Lorem ipsum.
		var variable

		# A function with long documentation. Lorem ipsum dolor sit amet.
		# Consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut
		# labore et dolore magna aliqua.
		fn function {|x| }
		`)
	fooVariableDoc = Dedent(`
		A variable. Lorem ipsum.
		`)
	// Function docs are used for checking the output of doc:show, so contains
	// the post-processed "Usage:" prefix.
	fooFunctionDoc = Dedent(`
		Usage:

		~~~elvish
		foo:function $x
		~~~

		A function with long documentation. Lorem ipsum dolor sit amet.
		Consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut
		labore et dolore magna aliqua.
		`)

	builtinModuleCode = Dedent(`
		# Terminates a loop.
		fn break { }

		# Constructs a [typed number](language.html#number). Another
		# [link](#foo).
		fn num {|x| }
		`)
	breakDoc = tildeToBackquote(Dedent(`
		Usage:

		~~~elvish
		break
		~~~

		Terminates a loop.
		`))
	// The relative link to language.html is converted to an absolute link.
	numDoc = Dedent(`
		Usage:

		~~~elvish
		num $x
		~~~

		Constructs a [typed number](https://elv.sh/ref/language.html#number).
		Another [link](#foo).
		`)
)

func setupDoc(ev *eval.Evaler) {
	ev.ExtendGlobal(eval.BuildNs().AddNs("doc", doc.Ns))
}

func render(s string, w int) string {
	return md.RenderString(s, &md.TTYCodec{Width: w, HighlightCodeBlock: doc.HighlightCodeBlock})
}
