package complete

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/elves/elvish/cli/addons/completion"
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
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
	return util.SetupTestDir(util.Dir{
		"exe":     util.File{Perm: 0755, Content: ""},
		"non-exe": "",
		"d": util.Dir{
			"exe": util.File{Perm: 0755, Content: ""},
		},
	}, "")
}

func TestComplete(t *testing.T) {
	cleanupFs := setupFs()
	defer cleanupFs()

	cfg := Config{
		Filter: PrefixFilter,
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
	}

	cfgWithCompleteArg := Config{
		PureEvaler: cfg.PureEvaler,
		CompleteArg: func(args []string) ([]RawItem, error) {
			item := noQuoteItem(fmt.Sprintf("%#v", args))
			return []RawItem{item}, nil
		},
	}

	allFileNameItems := []completion.Item{
		{ToShow: "d", ToInsert: "d/"},
		{ToShow: "exe", ToInsert: "exe "},
		{ToShow: "non-exe", ToInsert: "non-exe "},
	}

	allCommandItems := []completion.Item{
		c("bar = "), c("fn"), c("foo = "), c("for"), c("if"), c("ls"), c("make"),
		c("ns:"),
	}

	tt.Test(t, tt.Fn("Complete", Complete), tt.Table{
		// Complete arguments using the fallback filename completer.
		Args(cb("ls "), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(3, 3),
				Items: allFileNameItems},
			nil),
		Args(cb("ls e"), cfg).Rets(
			&Result{
				Name: "argument", Replace: r(3, 4),
				Items: []completion.Item{{ToShow: "exe", ToInsert: "exe "}}},
			nil),
		// Custom arg completer, new argument
		Args(cb("ls a "), cfgWithCompleteArg).Rets(
			&Result{
				Name: "argument", Replace: r(5, 5),
				Items: []completion.Item{c(`[]string{"ls", "a", ""}`)}},
			nil),
		Args(cb("ls a b"), cfgWithCompleteArg).Rets(
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

		// Complete local external commands.
		Args(cb("./"), cfg).Rets(
			&Result{
				Name: "command", Replace: r(0, 2),
				Items: []completion.Item{
					{ToShow: "./d", ToInsert: "./d/"},
					{ToShow: "./exe", ToInsert: "./exe "},
				}},
			nil),

		// TODO(xiaq): Add tests for completing indicies.

		// Complete filenames for redirection.
		Args(cb("p >"), cfg).Rets(
			&Result{Name: "redir", Replace: r(3, 3), Items: allFileNameItems},
			nil),
		Args(cb("p > e"), cfg).Rets(
			&Result{
				Name: "redir", Replace: r(4, 5),
				Items: []completion.Item{
					{ToShow: "exe", ToInsert: "exe "},
				}},
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

	// Symlinks are only available on UNIX.
	if goos := runtime.GOOS; goos != "windows" && goos != "plan9" {
		err := os.Symlink("d", "d2")
		if err != nil {
			panic(err)
		}
		tt.Test(t, tt.Fn("Complete", Complete), tt.Table{
			// Filename completion treats symlink to directories as directories.
			//       01234
			Args(cb("p > d"), cfg).Rets(
				&Result{
					Name: "redir", Replace: r(4, 5),
					Items: []completion.Item{
						{ToShow: "d", ToInsert: "d/"},
						{ToShow: "d2", ToInsert: "d2/"},
					}},
				nil,
			),
		})
	}
}

func cb(s string) CodeBuffer { return CodeBuffer{s, len(s)} }

func c(s string) completion.Item { return completion.Item{ToShow: s, ToInsert: s} }

func r(i, j int) diag.Ranging { return diag.Ranging{From: i, To: j} }
