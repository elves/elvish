package edit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
	"github.com/elves/getopt"
)

// completer takes the current Node (always a leaf in the AST) and an Editor,
// and should returns an interval and a list of candidates, meaning that the
// text within the interval may be replaced by any of the candidates. If the
// completer is not applicable, it should return an invalid interval [-1, end).
type completer func(parse.Node, *Editor) (int, int, []*candidate)

var completers = []struct {
	name string
	completer
}{
	{"variable", complVariable},
	{"command name", complFormHead},
	{"argument", complArg},
}

func complVariable(n parse.Node, ed *Editor) (int, int, []*candidate) {
	begin, end := n.Begin(), n.End()

	primary, ok := n.(*parse.Primary)
	if !ok || primary.Type != parse.Variable {
		return -1, -1, nil
	}

	splice, ns, head := eval.ParseVariable(primary.Value)

	// Collect matching variables.
	var varnames []string
	iterateVariables(ed.evaler, ns, func(varname string) {
		if strings.HasPrefix(varname, head) {
			varnames = append(varnames, varname)
		}
	})
	sort.Strings(varnames)

	cands := make([]*candidate, len(varnames))
	// Build candidates.
	for i, varname := range varnames {
		cands[i] = &candidate{text: "$" + eval.MakeVariableName(splice, ns, varname)}
	}
	return begin, end, cands
}

func iterateVariables(ev *eval.Evaler, ns string, f func(string)) {
	if ns == "" {
		for varname := range eval.Builtin() {
			f(varname)
		}
		for varname := range ev.Global {
			f(varname)
		}
		// TODO Include local names as well.
	} else {
		for varname := range ev.Modules[ns] {
			f(varname)
		}
	}
}

func complFormHead(n parse.Node, ed *Editor) (int, int, []*candidate) {
	begin, end, head, q := findFormHeadContext(n)
	if begin == -1 {
		return -1, -1, nil
	}
	cands, err := complFormHeadInner(head, ed)
	if err != nil {
		ed.Notify("%v", err)
	}
	fixCandidates(cands, q)
	return begin, end, cands
}

func findFormHeadContext(n parse.Node) (int, int, string, parse.PrimaryType) {
	_, isChunk := n.(*parse.Chunk)
	_, isPipeline := n.(*parse.Pipeline)
	if isChunk || isPipeline {
		return n.Begin(), n.End(), "", parse.Bareword
	}

	if primary, ok := n.(*parse.Primary); ok {
		if compound, head := primaryInSimpleCompound(primary); compound != nil {
			if form, ok := compound.Parent().(*parse.Form); ok {
				if form.Head == compound {
					return compound.Begin(), compound.End(), head, primary.Type
				}
			}
		}
	}
	return -1, -1, "", 0
}

func complFormHeadInner(head string, ed *Editor) ([]*candidate, error) {
	if util.DontSearch(head) {
		return complFilenameInner(head, true)
	}

	var commands []string
	got := func(s string) {
		if strings.HasPrefix(s, head) {
			commands = append(commands, s)
		}
	}
	for special := range isBuiltinSpecial {
		got(special)
	}
	splice, ns, _ := eval.ParseVariable(head)
	if !splice {
		iterateVariables(ed.evaler, ns, func(varname string) {
			if strings.HasPrefix(varname, eval.FnPrefix) {
				got(eval.MakeVariableName(false, ns, varname[len(eval.FnPrefix):]))
			} else {
				got(eval.MakeVariableName(false, ns, varname) + "=")
			}
		})
	}
	for command := range ed.isExternal {
		got(command)
		if strings.HasPrefix(head, "e:") {
			got("e:" + command)
		}
		if strings.HasPrefix(head, "E:") {
			got("E:" + command)
		}
	}
	sort.Strings(commands)

	cands := []*candidate{}
	for _, cmd := range commands {
		cands = append(cands, &candidate{text: cmd})
	}
	return cands, nil
}

func complArg(n parse.Node, ed *Editor) (int, int, []*candidate) {
	begin, end, current, q, form := findArgContext(n)
	if begin == -1 {
		return -1, -1, nil
	}

	// Find out head of the form and preceding arguments.
	// If Form.Head is not a simple compound, head will be "", just what we want.
	_, head, _ := simpleCompound(form.Head, nil)
	var args []string
	for _, compound := range form.Args {
		if compound.Begin() >= begin {
			break
		}
		ok, arg, _ := simpleCompound(compound, nil)
		if ok {
			// XXX Arguments that are not simple compounds are simply ignored.
			args = append(args, arg)
		}
	}

	words := make([]string, len(args)+2)
	words[0] = head
	words[len(words)-1] = current
	copy(words[1:len(words)-1], args[:])

	cands, err := completeArg(words, ed)
	if err != nil {
		ed.Notify("%v", err)
	}
	fixCandidates(cands, q)
	return begin, end, cands
}

func findArgContext(n parse.Node) (int, int, string, parse.PrimaryType, *parse.Form) {
	if sep, ok := n.(*parse.Sep); ok {
		if form, ok := sep.Parent().(*parse.Form); ok {
			return n.End(), n.End(), "", parse.Bareword, form
		}
	}
	if primary, ok := n.(*parse.Primary); ok {
		if compound, head := primaryInSimpleCompound(primary); compound != nil {
			if form, ok := compound.Parent().(*parse.Form); ok {
				if form.Head != compound {
					return compound.Begin(), compound.End(), head, primary.Type, form
				}
			}
		}
	}
	return -1, -1, "", 0, nil
}

// TODO: getStyle does redundant stats.
func complFilenameInner(head string, executableOnly bool) ([]*candidate, error) {
	dir, fileprefix := path.Split(head)
	if dir == "" {
		dir = "."
	}

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot list directory %s: %v", dir, err)
	}

	cands := []*candidate{}
	// Make candidates out of elements that match the file component.
	for _, info := range infos {
		name := info.Name()
		// Irrevelant file.
		if !strings.HasPrefix(name, fileprefix) {
			continue
		}
		// Hide dot files unless file starts with a dot.
		if !dotfile(fileprefix) && dotfile(name) {
			continue
		}
		// Only accept searchable directories and executable files if
		// executableOnly is true.
		if executableOnly && !(info.IsDir() || (info.Mode()&0111) != 0) {
			continue
		}

		// Full filename for source and getStyle.
		full := head + name[len(fileprefix):]

		suffix := " "
		if info.IsDir() {
			suffix = "/"
		} else if info.Mode()&os.ModeSymlink != 0 {
			stat, err := os.Stat(full)
			if err == nil && stat.IsDir() {
				// Symlink to directory.
				suffix = "/"
			}
		}

		cands = append(cands, &candidate{
			text: full, suffix: suffix,
			display: styled{name, defaultLsColor.getStyle(full)},
		})
	}

	return cands, nil
}

func fixCandidates(cands []*candidate, q parse.PrimaryType) []*candidate {
	for _, cand := range cands {
		quoted, _ := parse.QuoteAs(cand.text, q)
		cand.text = quoted + cand.suffix
	}
	return cands
}

func dotfile(fname string) bool {
	return strings.HasPrefix(fname, ".")
}

func complGetopt(ec *eval.EvalCtx, elemsv eval.IteratorValue, optsv eval.IteratorValue, argsv eval.IteratorValue) {
	var (
		elems    []string
		opts     []*getopt.Option
		args     []eval.FnValue
		variadic bool
	)
	desc := make(map[*getopt.Option]string)
	// Convert arguments.
	elemsv.Iterate(func(v eval.Value) bool {
		elem, ok := v.(eval.String)
		if !ok {
			throwf("arg should be string, got %s", v.Kind())
		}
		elems = append(elems, string(elem))
		return true
	})
	optsv.Iterate(func(v eval.Value) bool {
		m, ok := v.(eval.MapLike)
		if !ok {
			throwf("opt should be map-like, got %s", v.Kind())
		}
		opt := &getopt.Option{}
		vshort := maybeIndex(m, eval.String("short"))
		if vshort != nil {
			sv, ok := vshort.(eval.String)
			if !ok {
				throwf("short option should be string, got %s", vshort.Kind())
			}
			s := string(sv)
			r, size := utf8.DecodeRuneInString(s)
			if r == utf8.RuneError || size != len(s) {
				throwf("short option should be exactly one rune, got %v", parse.Quote(s))
			}
			opt.Short = r
		}
		vlong := maybeIndex(m, eval.String("long"))
		if vlong != nil {
			s, ok := vlong.(eval.String)
			if !ok {
				throwf("long option should be string, got %s", vlong.Kind())
			}
			opt.Long = string(s)
		}
		if vshort == nil && vlong == nil {
			throwf("opt should have at least one of short and long as keys")
		}
		vdesc := maybeIndex(m, eval.String("desc"))
		if vdesc != nil {
			s, ok := vdesc.(eval.String)
			if !ok {
				throwf("description must be string, got %s", vdesc.Kind())
			}
			desc[opt] = string(s)
		}
		opts = append(opts, opt)
		return true
	})
	argsv.Iterate(func(v eval.Value) bool {
		sv, ok := v.(eval.String)
		if ok {
			if string(sv) == "..." {
				variadic = true
				return true
			}
			throwf("string except for ... not allowed as argument handler, got %s", parse.Quote(string(sv)))
		}
		arg, ok := v.(eval.FnValue)
		if !ok {
			throwf("argument handler should be fn, got %s", v.Kind())
		}
		args = append(args, arg)
		return true
	})
	// TODO Configurable config
	g := getopt.Getopt{opts, getopt.GNUGetoptLong}
	_, parsedArgs, ctx := g.Parse(elems)
	out := ec.OutputChan()

	putShortOpt := func(opt *getopt.Option) {
		c := &candidate{text: "-" + string(opt.Short)}
		if d, ok := desc[opt]; ok {
			c.display.text = c.text + " (" + d + ")"
		}
		out <- c
	}
	putLongOpt := func(opt *getopt.Option) {
		c := &candidate{text: "--" + string(opt.Long)}
		if d, ok := desc[opt]; ok {
			c.display.text = c.text + " (" + d + ")"
		}
		out <- c
	}

	switch ctx.Type {
	case getopt.NewOptionOrArgument, getopt.Argument:
		// Find argument completer
		var argCompl eval.FnValue
		if len(parsedArgs) < len(args) {
			argCompl = args[len(parsedArgs)]
		} else if variadic {
			argCompl = args[len(args)-1]
		}
		if argCompl != nil {
			cands, err := callFnForCandidates(argCompl, ec.Evaler, []string{ctx.Text})
			maybeThrow(err)
			for _, cand := range cands {
				out <- cand
			}
		}
		// TODO Notify that there is no suitable argument completer
	case getopt.NewOption:
		for _, opt := range opts {
			if opt.Short != 0 {
				putShortOpt(opt)
			}
			if opt.Long != "" {
				putLongOpt(opt)
			}
		}
	case getopt.NewLongOption:
		for _, opt := range opts {
			if opt.Long != "" {
				putLongOpt(opt)
			}
		}
	case getopt.LongOption:
		for _, opt := range opts {
			if strings.HasPrefix(opt.Long, ctx.Text) {
				putLongOpt(opt)
			}
		}
	case getopt.ChainShortOption:
		for _, opt := range opts {
			if opt.Short != 0 {
				// XXX loses chained options
				putShortOpt(opt)
			}
		}
	case getopt.OptionArgument:
	}
}

func maybeIndex(m eval.MapLike, k eval.Value) eval.Value {
	if !m.HasKey(k) {
		return nil
	}
	return m.IndexOne(k)
}
