// Package evaltest supports testing the Elvish interpreter and libraries.
package evaltest

import (
	"bytes"
	"fmt"
	"go/build/constraint"
	"io/fs"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/diff"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/mods"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/transcript"
)

// TestTranscriptsInFS extracts all Elvish transcript sessions from .elv and
// .elvts files in fsys, and runs each of them as a test. See
// [src.elv.sh/pkg/transcript] for how transcript sessions are discovered.
//
// Typical use of this function looks like this:
//
//	import (
//		"embed"
//		"src.elv.sh/pkg/eval/evaltest"
//	)
//
//	//go:embed *.elv *.elvts
//	var transcripts embed.FS
//
//	func TestTranscripts(t *testing.T) {
//		evaltest.TestTranscriptsInFS(t, transcripts)
//	}
//
// The function accepts variadic arguments in (name, f) pairs, where name must
// not contain any spaces. Each pair defines a setup function that may be
// referred to in the transcripts with the directive "//name".
//
// The setup function f may take a *testing.T, *eval.Evaler and a string
// argument. All of them are optional but must appear in that order. If it takes
// a string argument, the directive can be followed by an argument after a space
// ("//name argument"), and that argument is passed to f. The argument itself
// may contain spaces.
//
// The following setup functions are predefined:
//
//   - in-temp-dir: Run inside a temporary directory.
//
//   - set-env $name $value: Run with the environment variable $name set to
//     $value.
//
//   - unset-env $name: Run with the environment variable $name unset.
//
//   - eval $code: Evaluate the argument as Elvish code.
//
//   - only-on $cond: Evaluate $cond like a //go:build constraint and only
//     run the test if the constraint is satisfied.
//
//     The full build constraint syntax is supported, but only literal GOARCH
//     and GOOS values and "unix" are recognized as tags; other tags are always
//     false.
//
//   - deprecation-level $x: Run with deprecation level set to $x.
//
// Since directives in a higher level propagated to all its descendants, this
// mechanism can be used to specify setup functions that apply to an entire
// .elvts file (or an entire elvish-transcript code block in a .elv file) or an
// entire section:
//
//	//global-setup
//
//	# h1 #
//	//h1-setup
//
//	## h2 ##
//	//h2-setup
//
//	// All of top-setup, h1-setup and h2-setup are run for this session, in that
//	// order.
//
//	~> echo foo
//	foo
func TestTranscriptsInFS(t *testing.T, fsys fs.FS, setupPairs ...any) {
	sessions, err := transcript.ParseFromFS(fsys)
	if err != nil {
		t.Fatalf("parse transcript sessions: %v", err)
	}
	testTranscripts(t, sessions, setupPairs)
}

func testTranscripts(t *testing.T, sessions []transcript.Session, setupPairs []any) {
	setupMap, argSetupMap := buildSetupMaps(setupPairs)
	for _, session := range sessions {
		t.Run(session.Name, func(t *testing.T) {
			ev := eval.NewEvaler()
			mods.AddTo(ev)
			for _, directive := range session.Directives {
				name, arg, _ := strings.Cut(directive, " ")
				if f, ok := setupMap[name]; ok {
					if arg != "" {
						t.Fatalf("setup function %s doesn't support arguments", name)
					}
					f(t, ev)
				} else if f, ok := argSetupMap[name]; ok {
					f(t, ev, arg)
				} else {
					t.Fatalf("unknown setup function: %s", name)
				}
			}
			for _, interaction := range session.Interactions {
				want := interaction.Output
				got := evalAndCollectOutput(t, ev, interaction.Code)
				if want != got {
					t.Errorf("\n%s\n-want +got:\n%s",
						interaction.PromptAndCode(), diff.DiffNoHeader(want, got))
				}
			}
		})
	}
}

func buildSetupMaps(setupPairs []any) (map[string]func(*testing.T, *eval.Evaler), map[string]func(*testing.T, *eval.Evaler, string)) {
	if len(setupPairs)%2 != 0 {
		panic(fmt.Sprintf("variadic arguments must come in pairs, got %d", len(setupPairs)))
	}
	setupMap := map[string]func(*testing.T, *eval.Evaler){
		"in-temp-dir": func(t *testing.T, ev *eval.Evaler) { testutil.InTempDir(t) },
	}
	argSetupMap := map[string]func(*testing.T, *eval.Evaler, string){
		"set-env": func(t *testing.T, ev *eval.Evaler, arg string) {
			name, value, _ := strings.Cut(arg, " ")
			testutil.Setenv(t, name, value)
		},
		"unset-env": func(t *testing.T, ev *eval.Evaler, name string) {
			testutil.Unsetenv(t, name)
		},
		"eval": func(t *testing.T, ev *eval.Evaler, code string) {
			err := ev.Eval(
				parse.Source{Name: "[setup]", Code: code},
				eval.EvalCfg{Ports: eval.DummyPorts})
			if err != nil {
				t.Fatalf("setup failed: %v\n", err)
			}
		},
		"only-on": func(t *testing.T, _ *eval.Evaler, arg string) {
			expr, err := constraint.Parse("//go:build " + arg)
			if err != nil {
				t.Fatal(err)
			}
			if !expr.Eval(func(tag string) bool {
				if tag == "unix" {
					return isUNIX
				}
				return tag == runtime.GOOS || tag == runtime.GOARCH
			}) {
				t.Skipf("constraint not satisfied: %s", arg)
			}
		},
		"deprecation-level": func(t *testing.T, _ *eval.Evaler, arg string) {
			testutil.Set(t, &prog.DeprecationLevel, must.OK1(strconv.Atoi(arg)))
		},
	}
	for i := 0; i < len(setupPairs); i += 2 {
		name := setupPairs[i].(string)
		if setupMap[name] != nil || argSetupMap[name] != nil {
			panic(fmt.Sprintf("there's already a setup functions named %s", name))
		}
		switch f := setupPairs[i+1].(type) {
		case func():
			setupMap[name] = func(_ *testing.T, _ *eval.Evaler) { f() }
		case func(*testing.T):
			setupMap[name] = func(t *testing.T, ev *eval.Evaler) { f(t) }
		case func(*eval.Evaler):
			setupMap[name] = func(t *testing.T, ev *eval.Evaler) { f(ev) }
		case func(*testing.T, *eval.Evaler):
			setupMap[name] = f
		case func(string):
			argSetupMap[name] = func(_ *testing.T, _ *eval.Evaler, s string) { f(s) }
		case func(*testing.T, string):
			argSetupMap[name] = func(t *testing.T, _ *eval.Evaler, s string) { f(t, s) }
		case func(*eval.Evaler, string):
			argSetupMap[name] = func(_ *testing.T, ev *eval.Evaler, s string) { f(ev, s) }
		case func(*testing.T, *eval.Evaler, string):
			argSetupMap[name] = f
		default:
			panic(fmt.Sprintf("unsupported setup function type: %T", f))
		}
	}
	return setupMap, argSetupMap
}

var valuePrefix = "▶ "

func evalAndCollectOutput(t *testing.T, ev *eval.Evaler, code string) string {
	port1, collect1 := must.OK2(eval.CapturePort())
	port2, collect2 := must.OK2(eval.CapturePort())
	ports := []*eval.Port{eval.DummyInputPort, port1, port2}

	ctx, done := eval.ListenInterrupts()
	err := ev.Eval(
		parse.Source{Name: "[tty]", Code: code},
		eval.EvalCfg{Ports: ports, Interrupts: ctx})
	done()

	values, stdout := collect1()
	_, stderr := collect2()

	var sb strings.Builder
	for _, value := range values {
		sb.WriteString(valuePrefix + vals.ReprPlain(value) + "\n")
	}
	sb.Write(normalizeLineEnding(stripSGR(stdout)))
	sb.Write(normalizeLineEnding(stripSGR(stderr)))

	if err != nil {
		if shower, ok := err.(diag.Shower); ok {
			sb.WriteString(stripSGRString(shower.Show("")))
		} else {
			sb.WriteString(err.Error())
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}

var sgrPattern = regexp.MustCompile("\033\\[[0-9;]*m")

func stripSGR(bs []byte) []byte      { return sgrPattern.ReplaceAllLiteral(bs, nil) }
func stripSGRString(s string) string { return sgrPattern.ReplaceAllLiteralString(s, "") }

func normalizeLineEnding(bs []byte) []byte { return bytes.ReplaceAll(bs, []byte("\r\n"), []byte("\n")) }
