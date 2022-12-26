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

		That("doc:show foo:bad").Throws(ErrorWithMessage("no doc for foo:bad")),
		That("doc:show bad:foo").Throws(ErrorWithMessage("no doc for bad:foo")),
	)
}

func TestSource(t *testing.T) {
	evaltest.TestWithSetup(t, setupDoc,
		That("doc:source foo:function").Puts(fooFunctionDoc),
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
	fooFunctionDoc = tildeToBackquote(Dedent(`
		~~~elvish
		foo:function $x
		~~~

		A function with long documentation. Lorem ipsum dolor sit amet,
		consectetur adipiscing elit, sed do eiusmod tempor incididunt ut
		labore et dolore magna aliqua.
		`))
	fooVariableDoc = Dedent(`
		A variable.
		`)

	builtinModuleCode = Dedent(`
		# Terminates a loop.
		fn break { }
		`)
	breakDoc = tildeToBackquote(Dedent(`
		~~~elvish
		break
		~~~

		Terminates a loop.
		`))
)

func setupDoc(ev *eval.Evaler) {
	ev.ExtendGlobal(eval.BuildNs().AddNs("doc", doc.Ns))
	doc.ModToCode["foo:"] = strings.NewReader(fooModuleCode)
	doc.ModToCode[""] = strings.NewReader(builtinModuleCode)
}

func render(s string, w int) string { return md.RenderString(s, &md.TTYCodec{Width: w}) }
