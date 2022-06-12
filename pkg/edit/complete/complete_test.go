package complete

// Mocked builtin commands

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"testing"

	"src.elv.sh/pkg/cli/lscolors"
	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/tt"
	"src.elv.sh/pkg/ui"
)

var Args = tt.Args

func TestComplete(t *testing.T) {
	lscolors.SetTestLsColors(t)
	testutil.InTempDir(t)
	testutil.ApplyDir(testutil.Dir{
		"a.exe":   testutil.File{Perm: 0755, Content: ""},
		"non-exe": "",
		"d": testutil.Dir{
			"a.exe": testutil.File{Perm: 0755, Content: ""},
		},
	})
	testutil.Set(t, &eachExternal, func(f func(string)) {
		f("external-cmd1")
		f("external-cmd2")
	})
	testutil.Set(t, &environ, func() []string {
		return []string{"ENV1=", "ENV2="}
	})

	ev := eval.NewEvaler()
	err := ev.Eval(parse.SourceForTest(strings.Join([]string{
		"var local-var1 = $nil",
		"var local-var2 = $nil",
		"fn local-fn1 { }",
		"fn local-fn2 { }",
		"var local-ns1: = (ns [&lorem=$nil])",
		"var local-ns2: = (ns [&ipsum=$nil])",
	}, "\n")), eval.EvalCfg{})
	if err != nil {
		t.Fatalf("evaler setup: %v", err)
	}
	ev.ReplaceBuiltin(
		eval.BuildNs().
			AddVar("builtin-var1", vars.NewReadOnly(nil)).
			AddVar("builtin-var2", vars.NewReadOnly(nil)).
			AddGoFn("builtin-fn1", func() {}).
			AddGoFn("builtin-fn2", func() {}).
			Ns())

	var cfg Config
	cfg = Config{
		Filterer: FilterPrefix,
		ArgGenerator: func(args []string) ([]RawItem, error) {
			if len(args) >= 2 && args[0] == "sudo" {
				return GenerateForSudo(args, ev, cfg)
			}
			return GenerateFileNames(args)
		},
	}

	argGeneratorDebugCfg := Config{
		Filterer: func(ctxName, seed string, items []RawItem) []RawItem {
			return items
		},
		ArgGenerator: func(args []string) ([]RawItem, error) {
			item := noQuoteItem(fmt.Sprintf("%#v", args))
			return []RawItem{item}, nil
		},
	}

	dupCfg := Config{
		ArgGenerator: func([]string) ([]RawItem, error) {
			return []RawItem{PlainItem("a"), PlainItem("b"), PlainItem("a")}, nil
		},
	}

	allFileNameItems := []modes.CompletionItem{
		fci("a.exe", " "), fci("d"+string(os.PathSeparator), ""), fci("non-exe", " "),
	}

	allCommandItems := []modes.CompletionItem{
		ci("builtin-fn1"), ci("builtin-fn2"),
		ci("external-cmd1"), ci("external-cmd2"),
		ci("local-fn1"), ci("local-fn2"),
		ci("local-ns1:"), ci("local-ns2:"),
	}
	// Add all special commands.
	for name := range eval.IsBuiltinSpecial {
		allCommandItems = append(allCommandItems, ci(name))
	}
	sort.Slice(allCommandItems, func(i, j int) bool {
		return allCommandItems[i].ToInsert < allCommandItems[j].ToInsert
	})

	tt.Test(t, tt.Fn("Complete", Complete), tt.Table{
		// Candidates are deduplicated.
		Args(cb("ls "), ev, dupCfg).Rets(
			&Result{
				Name: "argument", Replace: r(3, 3),
				Items: []modes.CompletionItem{
					ci("a"), ci("b"),
				},
			},
			nil),
		// Complete arguments using GenerateFileNames.
		Args(cb("ls "), ev, cfg).Rets(
			&Result{
				Name: "argument", Replace: r(3, 3),
				Items: allFileNameItems},
			nil),
		Args(cb("ls a"), ev, cfg).Rets(
			&Result{
				Name: "argument", Replace: r(3, 4),
				Items: []modes.CompletionItem{fci("a.exe", " ")}},
			nil),
		// GenerateForSudo completing external commands.
		Args(cb("sudo "), ev, cfg).Rets(
			&Result{
				Name: "argument", Replace: r(5, 5),
				Items: []modes.CompletionItem{ci("external-cmd1"), ci("external-cmd2")}},
			nil),
		// GenerateForSudo completing non-command arguments.
		Args(cb("sudo ls "), ev, cfg).Rets(
			&Result{
				Name: "argument", Replace: r(8, 8),
				Items: allFileNameItems},
			nil),
		// Custom arg completer, new argument
		Args(cb("ls a "), ev, argGeneratorDebugCfg).Rets(
			&Result{
				Name: "argument", Replace: r(5, 5),
				Items: []modes.CompletionItem{ci(`[]string{"ls", "a", ""}`)}},
			nil),
		Args(cb("ls a b"), ev, argGeneratorDebugCfg).Rets(
			&Result{
				Name: "argument", Replace: r(5, 6),
				Items: []modes.CompletionItem{ci(`[]string{"ls", "a", "b"}`)}},
			nil),

		// Complete for special command "set".
		Args(cb("set "), ev, cfg).Rets(
			&Result{
				Name: "argument", Replace: r(4, 4),
				Items: []modes.CompletionItem{
					ci("builtin-fn1~"), ci("builtin-fn2~"),
					ci("builtin-var1"), ci("builtin-var2"),
					ci("local-fn1~"), ci("local-fn2~"),
					ci("local-ns1:"), ci("local-ns2:"),
					ci("local-var1"), ci("local-var2"),
				},
			}),
		Args(cb("set @"), ev, cfg).Rets(
			&Result{
				Name: "argument", Replace: r(4, 5),
				Items: []modes.CompletionItem{
					ci("@builtin-fn1~"), ci("@builtin-fn2~"),
					ci("@builtin-var1"), ci("@builtin-var2"),
					ci("@local-fn1~"), ci("@local-fn2~"),
					ci("@local-ns1:"), ci("@local-ns2:"),
					ci("@local-var1"), ci("@local-var2"),
				},
			}),
		Args(cb("set local-ns1:"), ev, cfg).Rets(
			&Result{
				Name: "argument", Replace: r(4, 14),
				Items: []modes.CompletionItem{
					ci("local-ns1:lorem"),
				},
			}),
		Args(cb("set a = "), ev, cfg).Rets(
			&Result{
				Name: "argument", Replace: r(8, 8),
				Items: nil,
			}),
		// "tmp" has the same completer.
		Args(cb("tmp "), ev, cfg).Rets(
			&Result{
				Name: "argument", Replace: r(4, 4),
				Items: []modes.CompletionItem{
					ci("builtin-fn1~"), ci("builtin-fn2~"),
					ci("builtin-var1"), ci("builtin-var2"),
					ci("local-fn1~"), ci("local-fn2~"),
					ci("local-ns1:"), ci("local-ns2:"),
					ci("local-var1"), ci("local-var2"),
				},
			}),

		// Complete commands at an empty buffer, generating special forms,
		// externals, functions, namespaces and variable assignments.
		Args(cb(""), ev, cfg).Rets(
			&Result{Name: "command", Replace: r(0, 0), Items: allCommandItems},
			nil),
		// Complete at an empty closure.
		Args(cb("{ "), ev, cfg).Rets(
			&Result{Name: "command", Replace: r(2, 2), Items: allCommandItems},
			nil),
		// Complete after a newline.
		Args(cb("a\n"), ev, cfg).Rets(
			&Result{Name: "command", Replace: r(2, 2), Items: allCommandItems},
			nil),
		// Complete after a semicolon.
		Args(cb("a;"), ev, cfg).Rets(
			&Result{Name: "command", Replace: r(2, 2), Items: allCommandItems},
			nil),
		// Complete after a pipe.
		Args(cb("a|"), ev, cfg).Rets(
			&Result{Name: "command", Replace: r(2, 2), Items: allCommandItems},
			nil),
		// Complete at the beginning of output capture.
		Args(cb("a ("), ev, cfg).Rets(
			&Result{Name: "command", Replace: r(3, 3), Items: allCommandItems},
			nil),
		// Complete at the beginning of exception capture.
		Args(cb("a ?("), ev, cfg).Rets(
			&Result{Name: "command", Replace: r(4, 4), Items: allCommandItems},
			nil),

		// Complete external commands with the e: prefix.
		Args(cb("e:"), ev, cfg).Rets(
			&Result{
				Name: "command", Replace: r(0, 2),
				Items: []modes.CompletionItem{
					ci("e:external-cmd1"), ci("e:external-cmd2"),
				}},
			nil),

		// TODO(xiaq): Add tests for completing indices.

		// Complete filenames for redirection.
		Args(cb("p >"), ev, cfg).Rets(
			&Result{Name: "redir", Replace: r(3, 3), Items: allFileNameItems},
			nil),
		Args(cb("p > a"), ev, cfg).Rets(
			&Result{
				Name: "redir", Replace: r(4, 5),
				Items: []modes.CompletionItem{fci("a.exe", " ")}},
			nil),

		// Completing variables.
		Args(cb("p $"), ev, cfg).Rets(
			&Result{
				Name: "variable", Replace: r(3, 3),
				Items: []modes.CompletionItem{
					ci("E:"),
					ci("builtin-fn1~"), ci("builtin-fn2~"),
					ci("builtin-var1"), ci("builtin-var2"),
					ci("e:"),
					ci("local-fn1~"), ci("local-fn2~"),
					ci("local-ns1:"), ci("local-ns2:"),
					ci("local-var1"), ci("local-var2"),
				}},
			nil),
		Args(cb("p $local-"), ev, cfg).Rets(
			&Result{
				Name: "variable", Replace: r(3, 9),
				Items: []modes.CompletionItem{
					ci("local-fn1~"), ci("local-fn2~"),
					ci("local-ns1:"), ci("local-ns2:"),
					ci("local-var1"), ci("local-var2"),
				}},
			nil),
		//       01234567890123
		Args(cb("p $local-ns1:"), ev, cfg).Rets(
			&Result{
				Name: "variable", Replace: r(13, 13),
				Items: []modes.CompletionItem{ci("lorem")}},
			nil),
		//       012345
		Args(cb("p $e:"), ev, cfg).Rets(
			&Result{
				Name: "variable", Replace: r(5, 5),
				Items: []modes.CompletionItem{
					ci("external-cmd1~"), ci("external-cmd2~"),
				}},
			nil),
		//       012345
		Args(cb("p $E:"), ev, cfg).Rets(
			&Result{
				Name: "variable", Replace: r(5, 5),
				Items: []modes.CompletionItem{
					ci("ENV1"), ci("ENV2"),
				}},
			nil),
		//       01234567
		Args(cb("p $bad:"), ev, cfg).Rets(
			&Result{Name: "variable", Replace: r(7, 7)},
			nil),
		//       0123456789012345678901
		Args(cb("p $local-ns1:bad:bad:"), ev, cfg).Rets(
			&Result{Name: "variable", Replace: r(21, 21)},
			nil),

		// No completion in supported context.
		Args(cb("nop ["), ev, cfg).Rets((*Result)(nil), errNoCompletion),
		// No completion after parse error.
		Args(cb("nop `"), ev, cfg).Rets((*Result)(nil), errNoCompletion),
	})

	// Symlinks and executable bits are not available on Windows.
	if runtime.GOOS != "windows" {
		err := os.Symlink("d", "d2")
		if err != nil {
			panic(err)
		}
		allLocalCommandItems := []modes.CompletionItem{
			fci("./a.exe", " "), fci("./d/", ""), fci("./d2/", ""),
		}
		tt.Test(t, tt.Fn("Complete", Complete), tt.Table{
			// Filename completion treats symlink to directories as directories.
			//       01234
			Args(cb("p > d"), ev, cfg).Rets(
				&Result{
					Name: "redir", Replace: r(4, 5),
					Items: []modes.CompletionItem{fci("d/", ""), fci("d2/", "")}},
				nil,
			),

			// Complete local external commands.
			//
			// TODO(xiaq): Make this test applicable to Windows by using a
			// different criteria for executable files on Window.
			Args(cb("./"), ev, cfg).Rets(
				&Result{
					Name: "command", Replace: r(0, 2),
					Items: allLocalCommandItems},
				nil),
			// After sudo.
			Args(cb("sudo ./"), ev, cfg).Rets(
				&Result{
					Name: "argument", Replace: r(5, 7),
					Items: allLocalCommandItems},
				nil),
		})
	}
}

func cb(s string) CodeBuffer { return CodeBuffer{s, len(s)} }

func ci(s string) modes.CompletionItem { return modes.CompletionItem{ToShow: ui.T(s), ToInsert: s} }

func fci(s, suffix string) modes.CompletionItem {
	return modes.CompletionItem{
		ToShow:   ui.T(s, ui.StylingFromSGR(lscolors.GetColorist().GetStyle(s))),
		ToInsert: parse.Quote(s) + suffix}
}

func r(i, j int) diag.Ranging { return diag.Ranging{From: i, To: j} }
