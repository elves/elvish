// Package eval handles evaluation of nodes and consists the runtime of the
// shell.
package eval

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/xiaq/elvish/parse"
	"github.com/xiaq/elvish/util"
)

// Evaluator maintains runtime context of elvish code within a single
// goroutine. When elvish code spawns goroutines, the Evaluator is copied and
// has certain components replaced.
type Evaluator struct {
	checker     *Checker
	name, text  string
	globals     map[string]Value
	locals      map[string]Value
	env         *Env
	searchPaths []string
	in, out     *port
	statusCb    func([]string)
	nodes       []parse.Node // A stack that keeps track of nodes being evaluated.
}

func statusOk(ss []string) bool {
	for _, s := range ss {
		switch s {
		case "", "0":
		default:
			return false
		}
	}
	return true
}

// NewEvaluator creates a new Evaluator from a slice of environment strings
// in the form "key=value".
func NewEvaluator() *Evaluator {
	env := NewEnv()
	pid := NewString(strconv.Itoa(syscall.Getpid()))
	g := map[string]Value{
		"env": env, "pid": pid,
	}
	ev := &Evaluator{
		checker: NewChecker(),
		globals: g, locals: g, env: env,
		in: &port{f: os.Stdin}, out: &port{f: os.Stdout},
		statusCb: func(s []string) {
			if !statusOk(s) {
				fmt.Println("Status:", s)
			}
		},
	}

	path, ok := env.m["PATH"]
	if ok {
		ev.searchPaths = strings.Split(path, ":")
		// fmt.Printf("Search paths are %v\n", search_paths)
	} else {
		ev.searchPaths = []string{"/bin"}
	}

	return ev
}

func (ev *Evaluator) copy() *Evaluator {
	eu := new(Evaluator)
	*eu = *ev
	return eu
}

// Eval evaluates a chunk node n. The name and text of it is used for
// diagnostic messages.
func (ev *Evaluator) Eval(name, text string, n *parse.ChunkNode) (err error) {
	scope := make(map[string]bool)
	for name := range ev.globals {
		scope[name] = true
	}
	err = ev.checker.Check(name, text, n, scope)
	if err != nil {
		return
	}

	defer util.Recover(&err)
	defer ev.stopEval()
	ev.name = name
	ev.text = text
	ev.evalChunk(n)
	return nil
}

func (ev *Evaluator) stopEval() {
	ev.name = ""
	ev.text = ""
}

func (ev *Evaluator) push(n parse.Node) {
	ev.nodes = append(ev.nodes, n)
}

func (ev *Evaluator) pop() {
	n := len(ev.nodes) - 1
	ev.nodes[n] = nil
	ev.nodes = ev.nodes[:n]
}

func (ev *Evaluator) errorfNode(n parse.Node, format string, args ...interface{}) {
	util.Panic(util.NewContextualError(ev.name, ev.text, int(n.Position()), format, args...))
}

// errorf stops the evaluator. Its panic is supposed to be caught by recover.
func (ev *Evaluator) errorf(format string, args ...interface{}) {
	if n := len(ev.nodes); n > 0 {
		ev.errorfNode(ev.nodes[n-1], format, args...)
	} else {
		util.Panic(fmt.Errorf(format, args...))
	}
}

// ResolveVar tries to find an variable with the given name in the local and
// then global context of the Evaluator. If no variable with the name exists,
// err is non-nil.
func (ev *Evaluator) ResolveVar(name string) (v Value, err error) {
	defer util.Recover(&err)
	return ev.resolveVar(name), nil
}

func (ev *Evaluator) resolveVar(name string) Value {
	if val, ok := ev.locals[name]; ok {
		return val
	}
	if val, ok := ev.globals[name]; ok {
		return val
	}
	ev.errorf("Variable %q not found", name)
	return nil
}

func (ev *Evaluator) evalTable(tn *parse.TableNode) *Table {
	ev.push(tn)
	defer ev.pop()

	// Evaluate list part.
	t := NewTable()
	for _, term := range tn.List {
		vs := ev.evalTerm(term)
		t.append(vs...)
	}
	for _, pair := range tn.Dict {
		ks := ev.evalTerm(pair.Key)
		vs := ev.evalTerm(pair.Value)
		if len(ks) != len(vs) {
			ev.errorf("Number of keys doesn't match number of values: %d vs. %d", len(ks), len(vs))
		}
		for i, k := range ks {
			t.Dict[k] = vs[i]
		}
	}
	return t
}

func (ev *Evaluator) evalFactor(n *parse.FactorNode) []Value {
	ev.push(n)
	defer ev.pop()

	var words []Value

	switch n.Typ {
	case parse.StringFactor:
		m := n.Node.(*parse.StringNode)
		words = []Value{NewString(m.Text)}
	case parse.VariableFactor:
		m := n.Node.(*parse.StringNode)
		words = []Value{ev.resolveVar(m.Text)}
	case parse.ListFactor:
		m := n.Node.(*parse.TermListNode)
		words = ev.evalTermList(m)
	case parse.OutputCaptureFactor:
		m := n.Node.(*parse.PipelineNode)
		newEv := ev.copy()
		ch := make(chan Value)
		newEv.out = &port{ch: ch}
		updates := newEv.evalPipelineAsync(m)
		for v := range ch {
			words = append(words, v)
		}
		newEv.waitPipeline(updates)
	case parse.StatusCaptureFactor:
		m := n.Node.(*parse.PipelineNode)
		ss := ev.evalPipeline(m)
		words = make([]Value, len(ss))
		for i, s := range ss {
			words[i] = NewString(s)
		}
	case parse.TableFactor:
		m := n.Node.(*parse.TableNode)
		word := ev.evalTable(m)
		words = []Value{word}
	case parse.ClosureFactor:
		m := n.Node.(*parse.ClosureNode)
		var names []string
		if m.ArgNames != nil {
			nameValues := ev.evalTermList(m.ArgNames)
			for _, v := range nameValues {
				names = append(names, v.String(ev))
			}
		}
		words = []Value{NewClosure(names, m.Chunk)}
	default:
		panic("bad factor type")
	}

	return words
}

func (ev *Evaluator) evalTerm(n *parse.TermNode) []Value {
	ev.push(n)
	defer ev.pop()

	if len(n.Nodes) == 0 {
		panic("evalTerm got an empty list")
	}

	words := ev.evalFactor(n.Nodes[0])

	for _, m := range n.Nodes[1:] {
		a := ev.evalFactor(m)
		if len(a) == 1 {
			for i := range words {
				words[i] = words[i].Caret(ev, a[0])
			}
		} else {
			// Do a Cartesian product
			newWords := make([]Value, len(words)*len(a))
			for i := range words {
				for j := range a {
					newWords[i*len(a)+j] = words[i].Caret(ev, a[j])
				}
			}
			words = newWords
		}
	}
	return words
}

func (ev *Evaluator) evalTermList(ln *parse.TermListNode) []Value {
	ev.push(ln)
	defer ev.pop()

	words := make([]Value, 0, len(ln.Nodes))
	for _, n := range ln.Nodes {
		a := ev.evalTerm(n)
		words = append(words, a...)
	}
	return words
}

func (ev *Evaluator) asSingleString(vs []Value, n parse.Node, what string) *String {
	ev.push(n)
	defer ev.pop()

	if len(vs) != 1 {
		ev.errorf("Expect exactly one word for %s, got %d", what, len(vs))
	}
	v, ok := vs[0].(*String)
	if !ok {
		ev.errorf("Expect string for %s, got %s", what, vs[0])
	}
	return v
}

func (ev *Evaluator) evalTermSingleString(n *parse.TermNode, what string) *String {
	return ev.asSingleString(ev.evalTerm(n), n, what)
}

// BUG(xiaq): When evaluating a chunk, failure of one pipeline will abort the
// whole chunk.
func (ev *Evaluator) evalChunk(ch *parse.ChunkNode) {
	ev.push(ch)
	defer ev.pop()
	for _, n := range ch.Nodes {
		s := ev.evalPipeline(n)
		if ev.statusCb != nil {
			ev.statusCb(s)
		}
	}
}
