package complete

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/elves/elvish/cli/addons/completion"
	"github.com/elves/elvish/cli/lscolors"
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
	"github.com/elves/elvish/util"
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

func (ev testEvaler) PurelyEvalPartialCompound(
	cn *parse.Compound, upto *parse.Indexing) (string, error) {
	return (*eval.Evaler)(nil).PurelyEvalPartialCompound(cn, upto)
}

func (ev testEvaler) PurelyEvalCompound(cn *parse.Compound) (string, error) {
	return (*eval.Evaler)(nil).PurelyEvalCompound(cn)
}

func (ev testEvaler) PurelyEvalPrimary(pn *parse.Primary) interface{} {
	return (*eval.Evaler)(nil).PurelyEvalPrimary(pn)
}

func setupFs() func() {
	return util.InTestDirWithSetup(util.Dir{
		"a.exe":   util.File{Perm: 0755, Content: ""},
		"non-exe": "",
		"d": util.Dir{
			"a.exe": util.File{Perm: 0755, Content: ""},
		},
	})
}

func TestComplete(t *testing.T) {
	restoreLsColors := lscolors.WithTestLsColors()
	defer restoreLsColors()

	cleanupFs := setupFs()
	defer cleanupFs()

	var cfg Config
	cfg = Config{
		Filterer: FilterPrefix,
		PureEvaler: testEvaler{
			externals: []string{"ls", "make"},
			specials:  []string{"if", "for"},
			variables: map[string][]string{
				"":     []string{"foo", "bar", "fn~", "ns:"},
				"ns1:": []string{"lorem"},
				"ns2:": []string{"ipsum"},
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

	pathSep := parse.Quote(string(os.PathSeparator))
	allFileNameItems := []completion.Item{
		fc("a.exe", " "), fc("d", pathSep), fc("non-exe", " "),
	}

	allCommandItems := []completion.Item{
		c("bar = "), c("fn"), c("foo = "), c("for"), c("if"), c("ls"), c("make"),
		c("ns:"),
	}

	tt.Test(t, tt.Fn("Complete", Complete), tt.Table{
		// No PureEvaler.
		Args(cb(""), Config{}).Rets(
			(*Result)(nil),
			errNoPureEvaler),
		// Complete arguments using GenerateFileNames.
		Args(cb("ls "), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(3, 3),
				Items: allFileNameItems},
			nil),
		Args(cb("ls a"), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(3, 4),
				Items: []completion.Item{fc("a.exe", " ")}},
			nil),
		// GenerateForSudo completing external commands.
		Args(cb("sudo "), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(5, 5),
				Items: []completion.Item{c("ls"), c("make")}},
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
				Items: []completion.Item{c(`[]string{"ls", "a", ""}`)}},
			nil),
		Args(cb("ls a b"), argGeneratorDebugCfg).Rets(
			&Result{
				Name: "argument", Replace: r(5, 6),
				Items: []completion.Item{c(`[]string{"ls", "a", "b"}`)}},
			nil),

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
				Items: []completion.Item{c("e:ls"), c("e:make")}},
			nil),

		// TODO(xiaq): Add tests for completing indicies.

		// Complete filenames for redirection.
		Args(cb("p >"), cfg).Rets(
			&Result{Name: "redir", Replace: r(3, 3), Items: allFileNameItems},
			nil),
		Args(cb("p > a"), cfg).Rets(
			&Result{
				Name: "redir", Replace: r(4, 5),
				Items: []completion.Item{fc("a.exe", " ")}},
			nil),

		// Completing variables.
		Args(cb("p $"), cfg).Rets(
			&Result{
				Name: "variable", Replace: r(3, 3),
				Items: []completion.Item{
					c("bar"), c("fn~"), c("foo"), c("ns1:"), c("ns2:"), c("ns:")}},
			nil),
		Args(cb("p $f"), cfg).Rets(
			&Result{
				Name: "variable", Replace: r(3, 4),
				Items: []completion.Item{c("fn~"), c("foo")}},
			nil),
		//       0123456
		Args(cb("p $ns1:"), cfg).Rets(
			&Result{
				Name: "variable", Replace: r(7, 7),
				Items: []completion.Item{c("lorem")}},
			nil),
	})

	// Symlinks and executable bits are not available on Windows.
	if goos := runtime.GOOS; goos != "windows" {
		err := os.Symlink("d", "d2")
		if err != nil {
			panic(err)
		}
		allLocalCommandItems := []completion.Item{
			fc("./a.exe", " "), fc("./d", "/"), fc("./d2", "/"),
		}
		tt.Test(t, tt.Fn("Complete", Complete), tt.Table{
			// Filename completion treats symlink to directories as directories.
			//       01234
			Args(cb("p > d"), cfg).Rets(
				&Result{
					Name: "redir", Replace: r(4, 5),
					Items: []completion.Item{fc("d", "/"), fc("d2", "/")}},
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

func c(s string) completion.Item { return completion.Item{ToShow: s, ToInsert: s} }

func fc(s, suffix string) completion.Item {
	return completion.Item{ToShow: s, ToInsert: s + suffix,
		ShowStyle: styled.StyleFromSGR(lscolors.GetColorist().GetStyle(s))}
}

func r(i, j int) diag.Ranging { return diag.Ranging{From: i, To: j} }

func withPathSeparator(d string) string { return d + parse.Quote(string(os.PathSeparator)) }
