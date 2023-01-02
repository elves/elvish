package doc_test

import (
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

func TestShow(t *testing.T) {
	evaltest.TestWithSetup(t, setupDoc,
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

func TestSource(t *testing.T) {
	evaltest.TestWithSetup(t, setupDoc,
		That("doc:source '$foo:variable'").Puts(fooVariableDoc),
		// The implementation of doc:source is used by doc:show internally, so
		// we only test a simple case here.
	)
}

var tildeToBackquote = strings.NewReplacer("~", "`").Replace

var (
	fooModuleCode = Dedent(`
		# A variable.
		var variable

		# A function with long documentation. Lorem ipsum dolor sit amet,
		# consectetur adipiscing elit, sed do eiusmod tempor incididunt ut
		# labore et dolore magna aliqua.
		fn function {|x| }
		`)
	fooVariableDoc = Dedent(`
		A variable.
		`)
	// Function docs are used for checking the output of doc:show, so contains
	// the post-processed "Usage:" prefix.
	fooFunctionDoc = Dedent(`
		Usage:

		~~~elvish
		foo:function $x
		~~~

		A function with long documentation. Lorem ipsum dolor sit amet,
		consectetur adipiscing elit, sed do eiusmod tempor incididunt ut
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
	doc.ModToCode["foo:"] = strings.NewReader(fooModuleCode)
	doc.ModToCode[""] = strings.NewReader(builtinModuleCode)
}

func render(s string, w int) string {
	return md.RenderString(s, &md.TTYCodec{Width: w, HighlightCodeBlock: doc.HighlightCodeBlock})
}
