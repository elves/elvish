// Package evaltest supports testing the Elvish interpreter and libraries.
//
// The entrypoint of this package is [TestTranscriptsInFS]. Typical usage looks
// like this:
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
// See [src.elv.sh/pkg/transcript] for how transcript sessions are discovered.
//
// # Setup functions
//
// [TestTranscriptsInFS] accepts variadic arguments in (name, f) pairs, where
// name must not contain any spaces. Each pair defines a setup function that may
// be referred to in the transcripts with the directive "//name".
//
// The setup function f may take a [*testing.T], [*eval.Evaler] and a string
// argument. All of them are optional but must appear in that order. If it takes
// a string argument, the directive can be followed by an argument after a space
// ("//name argument"), and that argument is passed to f. The argument itself
// may contain spaces.
//
// The following setup functions are predefined:
//
//   - skip-test: Don't run this test. Useful for examples in .d.elv files that
//     shouldn't be run as tests.
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
//     The syntax is the same as //go:build constraints, but the set of
//     supported tags is different and consists of: GOARCH and GOOS values,
//     "unix", "32bit" and "64bit".
//
//   - deprecation-level $x: Run with deprecation level set to $x.
//
// These setup functions can then be used in transcripts as directives. By
// default, they only apply to the current session; adding a "each:" prefix
// makes them apply to descendant sessions too.
//
//	//global-setup
//	//each:global-setup-2
//
//	# h1 #
//	//h1-setup
//	//each:h1-setup2
//
//	## h2 ##
//	//h2-setup
//
//	// All of globa-setup2, h1-setup2 and h2-setup are run for this session, in
//	// that
//
//	~> echo foo
//	foo
//
// # ELVISH_TRANSCRIPT_RUN
//
// The environment variable ELVISH_TRANSCRIPT_RUN may be set to a string
// $filename:$lineno. If the location falls within the code lines of an
// interaction, the following happens:
//
//  1. Only the session that the interaction belongs to is run, and only up to
//     the located interaction.
//
//  2. If the actual output doesn't match what's in the file, the test fails,
//     and writes out a machine readable instruction to update the file to match
//     the actual output.
//
// As an example, consider the following fragment of foo_test.elvts (with line
// numbers):
//
//	12 ~> echo foo
//	13    echo bar
//	14 lorem
//	15 ipsum
//
// Running
//
//	env ELVISH_TRANSCRIPT_RUN=foo_test.elvts:12 go test -run TestTranscripts
//
// will end up with a test failure, with a message like the following (the line
// range is left-closed, right-open):
//
//	UPDATE {"fromLine": 14, "toLine": 16, "content": "foo\nbar\n"}
//
// This mechanism enables editor plugins that can fill or update the output of
// transcript tests without requiring user to leave the editor.
//
// # Deterministic output order
//
// When Elvish code writes to both the value output and byte output, or to both
// stdout and stderr, there's no guarantee which one appears first in the
// terminal.
//
// To make testing easier, this package guarantees a deterministic order in such
// cases.
package evaltest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/build/constraint"
	"io/fs"
	"math"
	"os"
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
// .elvts files in fsys, and runs each of them as a test.
func TestTranscriptsInFS(t *testing.T, fsys fs.FS, setupPairs ...any) {
	nodes, err := transcript.ParseFromFS(fsys)
	if err != nil {
		t.Fatalf("parse transcript sessions: %v", err)
	}
	TestTranscriptNodes(t, nodes, setupPairs...)
}

// TestTranscriptsInFS runs parsed Elvish transcript nodes as tests.
func TestTranscriptNodes(t *testing.T, nodes []*transcript.Node, setupPairs ...any) {
	var run *runCfg
	if runEnv := os.Getenv("ELVISH_TRANSCRIPT_RUN"); runEnv != "" {
		filename, lineNo, ok := parseFileNameAndLineNo(runEnv)
		if !ok {
			t.Fatalf("can't parse ELVISH_TRANSCRIPT_RUN: %q", runEnv)
		}
		var node *transcript.Node
		for _, n := range nodes {
			if n.Name == filename {
				node = n
				break
			}
		}
		if node == nil {
			t.Fatalf("can't find file %q", filename)
		}
		nodes = []*transcript.Node{node}
		outputPrefix := ""
		if strings.HasSuffix(filename, ".elv") {
			outputPrefix = "# "
		}
		run = &runCfg{lineNo, outputPrefix}
	}
	testTranscripts(t, buildSetupDirectives(setupPairs), nodes, nil, run)
}

type runCfg struct {
	line         int
	outputPrefix string
}

func parseFileNameAndLineNo(s string) (string, int, bool) {
	i := strings.LastIndexByte(s, ':')
	if i == -1 {
		return "", 0, false
	}
	filename, lineNoString := s[:i], s[i+1:]
	lineNo, err := strconv.Atoi(lineNoString)
	if err != nil {
		return "", 0, false
	}
	return filename, lineNo, true
}

var solPattern = regexp.MustCompile("(?m:^)")

func testTranscripts(t *testing.T, sd *setupDirectives, nodes []*transcript.Node, setups []setupFunc, run *runCfg) {
	for _, node := range nodes {
		if run != nil && !(node.LineFrom <= run.line && run.line < node.LineTo) {
			continue
		}
		t.Run(node.Name, func(t *testing.T) {
			ev := eval.NewEvaler()
			mods.AddTo(ev)
			for _, setup := range setups {
				setup(t, ev)
			}
			var eachSetups []setupFunc
			for _, directive := range node.Directives {
				setup, each, err := sd.compile(directive)
				if err != nil {
					t.Fatal(err)
				}
				setup(t, ev)
				if each {
					eachSetups = append(eachSetups, setup)
				}
			}
			for _, interaction := range node.Interactions {
				if run != nil && interaction.CodeLineFrom > run.line {
					break
				}
				want := interaction.Output
				got := evalAndCollectOutput(ev, interaction.Code)
				if want != got {
					if run == nil {
						t.Errorf("\n%s\n-want +got:\n%s",
							interaction.PromptAndCode(), diff.DiffNoHeader(want, got))
					} else if interaction.CodeLineFrom <= run.line && run.line < interaction.CodeLineTo {
						content := got
						if run.outputPrefix != "" {
							// Insert output prefix at each SOL, except for the
							// SOL after the trailing newline.
							content = solPattern.ReplaceAllLiteralString(strings.TrimSuffix(content, "\n"), run.outputPrefix) + "\n"
						}
						correction := struct {
							FromLine int    `json:"fromLine"`
							ToLine   int    `json:"toLine"`
							Content  string `json:"content"`
						}{interaction.OutputLineFrom, interaction.OutputLineTo, content}
						t.Errorf("UPDATE %s", must.OK1(json.Marshal(correction)))
					}
				}
			}
			if len(node.Children) > 0 {
				// TODO: Use slices.Concat when Elvish requires Go 1.22
				allSetups := make([]setupFunc, 0, len(setups)+len(eachSetups))
				allSetups = append(allSetups, setups...)
				allSetups = append(allSetups, eachSetups...)
				testTranscripts(t, sd, node.Children, allSetups, run)
			}
		})
	}
}

type (
	setupFunc    func(*testing.T, *eval.Evaler)
	argSetupFunc func(*testing.T, *eval.Evaler, string)
)

type setupDirectives struct {
	setupMap    map[string]setupFunc
	argSetupMap map[string]argSetupFunc
}

func buildSetupDirectives(setupPairs []any) *setupDirectives {
	if len(setupPairs)%2 != 0 {
		panic(fmt.Sprintf("variadic arguments must come in pairs, got %d", len(setupPairs)))
	}
	setupMap := map[string]setupFunc{
		"in-temp-dir": func(t *testing.T, ev *eval.Evaler) { testutil.InTempDir(t) },
		"skip-test":   func(t *testing.T, _ *eval.Evaler) { t.SkipNow() },
	}
	argSetupMap := map[string]argSetupFunc{
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
				t.Fatalf("parse constraint %q: %v", arg, err)
			}
			if !expr.Eval(func(tag string) bool {
				switch tag {
				case "unix":
					return isUNIX
				case "32bit":
					return math.MaxInt == math.MaxInt32
				case "64bit":
					return math.MaxInt == math.MaxInt64
				default:
					return tag == runtime.GOOS || tag == runtime.GOARCH
				}
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
	return &setupDirectives{setupMap, argSetupMap}
}

func (sd *setupDirectives) compile(directive string) (f setupFunc, each bool, err error) {
	cutDirective := directive
	if s, ok := strings.CutPrefix(directive, "each:"); ok {
		cutDirective = s
		each = true
	}
	name, arg, _ := strings.Cut(cutDirective, " ")
	if f, ok := sd.setupMap[name]; ok {
		if arg != "" {
			return nil, false, fmt.Errorf("setup function %q doesn't support arguments", name)
		}
		return f, each, nil
	} else if f, ok := sd.argSetupMap[name]; ok {
		return func(t *testing.T, ev *eval.Evaler) {
			f(t, ev, arg)
		}, each, nil
	} else {
		return nil, false, fmt.Errorf("unknown setup function %q in directive %q", name, directive)
	}
}

var valuePrefix = "▶ "

func evalAndCollectOutput(ev *eval.Evaler, code string) string {
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
