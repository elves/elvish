package complete

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"src.elv.sh/pkg/cli/lscolors"
	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/tt"
	"src.elv.sh/pkg/ui"
)

var Args = tt.Args

// An implementation of PureEvaler useful in tests.
type testEvaler struct {
	externals  []string
	specials   []string
	namespaces []string
	variables  map[string][]string
}

func feed(f func(string), ss []string) {
	for _, s := range ss {
		f(s)
	}
}

func (ev testEvaler) EachExternal(f func(string)) { feed(f, ev.externals) }
func (ev testEvaler) EachSpecial(f func(string))  { feed(f, ev.specials) }
func (ev testEvaler) EachNs(f func(string))       { feed(f, ev.namespaces) }

func (ev testEvaler) EachVariableInNs(ns string, f func(string)) {
	feed(f, ev.variables[ns])
}

func (ev testEvaler) PurelyEvalPartialCompound(cn *parse.Compound, upto int) (string, bool) {
	return (*eval.Evaler)(nil).PurelyEvalPartialCompound(cn, upto)
}

func (ev testEvaler) PurelyEvalCompound(cn *parse.Compound) (string, bool) {
	return (*eval.Evaler)(nil).PurelyEvalCompound(cn)
}

func (ev testEvaler) PurelyEvalPrimary(pn *parse.Primary) interface{} {
	return (*eval.Evaler)(nil).PurelyEvalPrimary(pn)
}

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

	var cfg Config
	cfg = Config{
		Filterer: FilterPrefix,
		PureEvaler: testEvaler{
			externals: []string{"ls", "make"},
			specials:  []string{"if", "for"},
			variables: map[string][]string{
				"":     {"foo", "bar", "fn~", "ns:"},
				"ns1:": {"lorem"},
				"ns2:": {"ipsum"},
			},
			namespaces: []string{"ns1:", "ns2:"},
		},
		ArgGenerator: func(args []string) ([]RawItem, error) {
			if len(args) >= 2 && args[0] == "sudo" {
				return GenerateForSudo(cfg, args)
			}
			return GenerateFileNames(args)
		},
	}

	argGeneratorDebugCfg := Config{
		PureEvaler: cfg.PureEvaler,
		Filterer: func(ctxName, seed string, items []RawItem) []RawItem {
			return items
		},
		ArgGenerator: func(args []string) ([]RawItem, error) {
			item := noQuoteItem(fmt.Sprintf("%#v", args))
			return []RawItem{item}, nil
		},
	}

	dupCfg := Config{
		PureEvaler: cfg.PureEvaler,
		ArgGenerator: func([]string) ([]RawItem, error) {
			return []RawItem{PlainItem("a"), PlainItem("b"), PlainItem("a")}, nil
		},
	}

	allFileNameItems := []modes.CompletionItem{
		fc("a.exe", " "), fc("d"+string(os.PathSeparator), ""), fc("non-exe", " "),
	}

	allCommandItems := []modes.CompletionItem{
		c("fn"), c("for"), c("if"), c("ls"), c("make"), c("ns:"),
	}

	tt.Test(t, tt.Fn("Complete", Complete), tt.Table{
		// No PureEvaler.
		Args(cb(""), Config{}).Rets(
			(*Result)(nil),
			errNoPureEvaler),
		// Candidates are deduplicated.
		Args(cb("ls "), dupCfg).Rets(
			&Result{
				Name: "argument", Replace: r(3, 3),
				Items: []modes.CompletionItem{
					c("a"), c("b"),
				},
			},
			nil),
		// Complete arguments using GenerateFileNames.
		Args(cb("ls "), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(3, 3),
				Items: allFileNameItems},
			nil),
		Args(cb("ls a"), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(3, 4),
				Items: []modes.CompletionItem{fc("a.exe", " ")}},
			nil),
		// GenerateForSudo completing external commands.
		Args(cb("sudo "), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(5, 5),
				Items: []modes.CompletionItem{c("ls"), c("make")}},
			nil),
		// GenerateForSudo completing non-command arguments.
		Args(cb("sudo ls "), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(8, 8),
				Items: allFileNameItems},
			nil),
		// Custom arg completer, new argument
		Args(cb("ls a "), argGeneratorDebugCfg).Rets(
			&Result{
				Name: "argument", Replace: r(5, 5),
				Items: []modes.CompletionItem{c(`[]string{"ls", "a", ""}`)}},
			nil),
		Args(cb("ls a b"), argGeneratorDebugCfg).Rets(
			&Result{
				Name: "argument", Replace: r(5, 6),
				Items: []modes.CompletionItem{c(`[]string{"ls", "a", "b"}`)}},
			nil),

		// Complete for special command "set".
		Args(cb("set "), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(4, 4),
				Items: []modes.CompletionItem{
					c("bar"), c("fn~"), c("foo"), c("ns:"),
				},
			}),
		Args(cb("set @"), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(4, 5),
				Items: []modes.CompletionItem{
					c("@bar"), c("@fn~"), c("@foo"), c("@ns:"),
				},
			}),
		Args(cb("set ns1:"), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(4, 8),
				Items: []modes.CompletionItem{
					c("ns1:lorem"),
				},
			}),
		Args(cb("set a = "), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(8, 8),
				Items: nil,
			}),
		// "tmp" has the same completer.
		Args(cb("tmp "), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(4, 4),
				Items: []modes.CompletionItem{
					c("bar"), c("fn~"), c("foo"), c("ns:"),
				},
			}),

		// Complete commands at an empty buffer, generating special forms,
		// externals, functions, namespaces and variable assignments.
		Args(cb(""), cfg).Rets(
			&Result{Name: "command", Replace: r(0, 0), Items: allCommandItems},
			nil),
		// Complete at an empty closure.
		Args(cb("{ "), cfg).Rets(
			&Result{Name: "command", Replace: r(2, 2), Items: allCommandItems},
			nil),
		// Complete after a newline.
		Args(cb("a\n"), cfg).Rets(
			&Result{Name: "command", Replace: r(2, 2), Items: allCommandItems},
			nil),
		// Complete after a semicolon.
		Args(cb("a;"), cfg).Rets(
			&Result{Name: "command", Replace: r(2, 2), Items: allCommandItems},
			nil),
		// Complete after a pipe.
		Args(cb("a|"), cfg).Rets(
			&Result{Name: "command", Replace: r(2, 2), Items: allCommandItems},
			nil),
		// Complete at the beginning of output capture.
		Args(cb("a ("), cfg).Rets(
			&Result{Name: "command", Replace: r(3, 3), Items: allCommandItems},
			nil),
		// Complete at the beginning of exception capture.
		Args(cb("a ?("), cfg).Rets(
			&Result{Name: "command", Replace: r(4, 4), Items: allCommandItems},
			nil),

		// Complete external commands with the e: prefix.
		Args(cb("e:"), cfg).Rets(
			&Result{
				Name: "command", Replace: r(0, 2),
				Items: []modes.CompletionItem{c("e:ls"), c("e:make")}},
			nil),

		// TODO(xiaq): Add tests for completing indices.

		// Complete filenames for redirection.
		Args(cb("p >"), cfg).Rets(
			&Result{Name: "redir", Replace: r(3, 3), Items: allFileNameItems},
			nil),
		Args(cb("p > a"), cfg).Rets(
			&Result{
				Name: "redir", Replace: r(4, 5),
				Items: []modes.CompletionItem{fc("a.exe", " ")}},
			nil),

		// Completing variables.
		Args(cb("p $"), cfg).Rets(
			&Result{
				Name: "variable", Replace: r(3, 3),
				Items: []modes.CompletionItem{
					c("bar"), c("fn~"), c("foo"), c("ns1:"), c("ns2:"), c("ns:")}},
			nil),
		Args(cb("p $f"), cfg).Rets(
			&Result{
				Name: "variable", Replace: r(3, 4),
				Items: []modes.CompletionItem{c("fn~"), c("foo")}},
			nil),
		//       0123456
		Args(cb("p $ns1:"), cfg).Rets(
			&Result{
				Name: "variable", Replace: r(7, 7),
				Items: []modes.CompletionItem{c("lorem")}},
			nil),
	})

	// Symlinks and executable bits are not available on Windows.
	if goos := runtime.GOOS; goos != "windows" {
		err := os.Symlink("d", "d2")
		if err != nil {
			panic(err)
		}
		allLocalCommandItems := []modes.CompletionItem{
			fc("./a.exe", " "), fc("./d/", ""), fc("./d2/", ""),
		}
		tt.Test(t, tt.Fn("Complete", Complete), tt.Table{
			// Filename completion treats symlink to directories as directories.
			//       01234
			Args(cb("p > d"), cfg).Rets(
				&Result{
					Name: "redir", Replace: r(4, 5),
					Items: []modes.CompletionItem{fc("d/", ""), fc("d2/", "")}},
				nil,
			),

			// Complete local external commands.
			//
			// TODO(xiaq): Make this test applicable to Windows by using a
			// different criteria for executable files on Window.
			Args(cb("./"), cfg).Rets(
				&Result{
					Name: "command", Replace: r(0, 2),
					Items: allLocalCommandItems},
				nil),
			// After sudo.
			Args(cb("sudo ./"), cfg).Rets(
				&Result{
					Name: "argument", Replace: r(5, 7),
					Items: allLocalCommandItems},
				nil),
		})
	}
}

func cb(s string) CodeBuffer { return CodeBuffer{s, len(s)} }

func c(s string) modes.CompletionItem { return modes.CompletionItem{ToShow: ui.T(s), ToInsert: s} }

func fc(s, suffix string) modes.CompletionItem {
	return modes.CompletionItem{
		ToShow:   ui.T(s, ui.StylingFromSGR(lscolors.GetColorist().GetStyle(s))),
		ToInsert: parse.Quote(s) + suffix}
}

func r(i, j int) diag.Ranging { return diag.Ranging{From: i, To: j} }
