package shell

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/prog"
	. "src.elv.sh/pkg/prog/progtest"
)

func TestMain(m *testing.M) {
	interactiveRescueShell = false
	os.Exit(m.Run())
}

func TestInteract_SingleCommand(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("echo hello\n")

	Interact(f.Fds(), &InteractConfig{})
	f.TestOut(t, 1, "hello\n")
}

func TestInteract_Exception(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("fail mock\n")

	Interact(f.Fds(), &InteractConfig{})
	f.TestOutSnippet(t, 2, "fail mock")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	MustWriteFile("rc.elv", "echo hello from rc.elv")

	Interact(f.Fds(), &InteractConfig{Paths: Paths{Rc: "rc.elv"}})
	f.TestOut(t, 1, "hello from rc.elv\n")
}

func TestInteract_RcFile_DoesNotCompile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	MustWriteFile("rc.elv", "echo $a")

	Interact(f.Fds(), &InteractConfig{Paths: Paths{Rc: "rc.elv"}})
	f.TestOutSnippet(t, 2, "variable $a not found")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile_Exception(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	MustWriteFile("rc.elv", "fail mock")

	Interact(f.Fds(), &InteractConfig{Paths: Paths{Rc: "rc.elv"}})
	f.TestOutSnippet(t, 2, "fail mock")
	f.TestOut(t, 1, "")
}

func TestInteract_RcFile_NonexistentIsOK(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	f.FeedIn("")

	Interact(f.Fds(), &InteractConfig{Paths: Paths{Rc: "rc.elv"}})
	f.TestOut(t, 1, "")
}

func TestExtractExports(t *testing.T) {
	restore := prog.SetDeprecationLevel(0)
	defer restore()

	ns := eval.NsBuilder{
		exportsVarName: vars.NewReadOnly(exportsVarName, vals.EmptyMap.Assoc("a", "lorem")),
	}.Ns()
	ns2 := extractExports(ns, &bytes.Buffer{})
	if a, _ := ns2.Index("a"); a != "lorem" {
		t.Errorf("$a not extracted from exports")
	}
}

func TestExtractExports_IgnoreNonMapExports(t *testing.T) {
	restore := prog.SetDeprecationLevel(0)
	defer restore()

	ns := eval.NsBuilder{
		exportsVarName: vars.NewReadOnly(exportsVarName, "x"),
	}.Ns()
	var errBuf bytes.Buffer
	extractExports(ns, &errBuf)
	if errBuf.Len() == 0 {
		t.Errorf("No error written with non-map exports")
	}
}

func TestExtractExports_IgnoreNonStringKeys(t *testing.T) {
	restore := prog.SetDeprecationLevel(0)
	defer restore()

	ns := eval.NsBuilder{
		exportsVarName: vars.NewReadOnly(exportsVarName, vals.EmptyMap.Assoc(vals.EmptyList, "lorem")),
	}.Ns()
	var errBuf bytes.Buffer
	extractExports(ns, &errBuf)
	if errBuf.Len() == 0 {
		t.Errorf("No error written with non-string key")
	}
}

func TestExtractExports_DoesNotOverwrite(t *testing.T) {
	restore := prog.SetDeprecationLevel(0)
	defer restore()

	ns := eval.NsBuilder{
		"a":            vars.NewReadOnly("a", "lorem"),
		exportsVarName: vars.NewReadOnly(exportsVarName, vals.EmptyMap.Assoc("a", "ipsum")),
	}.Ns()
	var errBuf bytes.Buffer
	ns2 := extractExports(ns, &errBuf)
	if ns2.HasName("a") {
		t.Errorf("Existing variable overwritten")
	}
	if errBuf.Len() == 0 {
		t.Errorf("No error written with name conflict")
	}
}

func TestExtractExports_ShowsDeprecationWarning(t *testing.T) {
	restore := prog.SetDeprecationLevel(15)
	defer restore()

	ns := eval.NsBuilder{
		exportsVarName: vars.NewReadOnly(exportsVarName, vals.EmptyMap),
	}.Ns()
	var errBuf bytes.Buffer
	extractExports(ns, &errBuf)
	if errOut := errBuf.String(); !strings.Contains(errOut, "deprecated") {
		t.Errorf("no deprecation warning")
	}
}
