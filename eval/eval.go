// Package eval handles evaluation of nodes and consists the runtime of the
// shell.
package eval

import (
	"os"
	"fmt"
	"strings"
	"syscall"
	"strconv"
	"../parse"
	"../util"
)

type Evaluator struct {
	name, text string
	globals map[string]Value
	locals map[string]Value
	env *Env
	searchPaths []string
	filesToClose []*os.File
	nodes []parse.Node // A stack that keeps track of nodes being evaluated.
}

func NewEvaluator(envSlice []string) *Evaluator {
	env := NewEnv(envSlice)
	pid := NewScalar(strconv.Itoa(syscall.Getpid()))
	g := map[string]Value{
		"env": env, "pid": pid,
	}
	ev := &Evaluator{globals: g, locals: g, env: env}

	path, ok := env.m["PATH"]
	if ok {
		ev.searchPaths = strings.Split(path, ":")
		// fmt.Printf("Search paths are %v\n", search_paths)
	} else {
		ev.searchPaths = []string{"/bin"}
	}

	return ev
}

func (ev *Evaluator) Eval(name, text string, n parse.Node) (err error) {
	defer util.Recover(&err)
	defer ev.stopEval()
	ev.name = name
	ev.text = text
	ev.evalChunk(n.(*parse.ListNode))
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
func (ev *Evaluator) errorf(format string, args...interface{}) {
	ev.errorfNode(ev.nodes[len(ev.nodes) - 1], format, args...)
}

// XXX Duplicate with resolveVar
func (ev *Evaluator) getVar(name string) (Value, bool) {
	if val, ok := ev.locals[name]; ok {
		return val, true
	}
	if val, ok := ev.globals[name]; ok {
		return val, true
	}
	return nil, false
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
			t.dict[k] = vs[i]
		}
	}
	return t
}

func (ev *Evaluator) evalFactor(n *parse.FactorNode) []Value {
	ev.push(n)
	defer ev.pop()

	var words []Value

	switch n := n.Node.(type) {
	case *parse.StringNode:
		words = []Value{NewScalar(n.Text)}
	case *parse.ListNode:
		words = ev.evalTermList(n)
	case *parse.TableNode:
		word := ev.evalTable(n)
		words = []Value{word}
	case *parse.ClosureNode:
		var names []string
		if n.ArgNames != nil {
			nameValues := ev.evalTermList(n.ArgNames)
			for _, v := range nameValues {
				names = append(names, v.String(ev))
			}
		}
		words = []Value{NewClosure(names, n.Chunk)}
	default:
		panic("bad node type")
	}

	for dollar := n.Dollar; dollar > 0; dollar-- {
		if len(words) != 1 {
			ev.errorf("Only a single value may be dollared")
		}
		if _, ok := words[0].(*Scalar); !ok {
			ev.errorf("Only scalar may be dollared (for now)")
		}
		words[0] = ev.resolveVar(words[0].(*Scalar).str)
	}

	return words
}

func (ev *Evaluator) evalTerm(n *parse.ListNode) []Value {
	ev.push(n)
	defer ev.pop()

	if len(n.Nodes) == 0 {
		panic("evalTerm got an empty list")
	}

	words := ev.evalFactor(n.Nodes[0].(*parse.FactorNode))

	for _, m := range n.Nodes[1:] {
		a := ev.evalFactor(m.(*parse.FactorNode))
		if len(a) == 1 {
			for i := range words {
				words[i] = words[i].Caret(ev, a[0])
			}
		} else {
			// Do a Cartesian product
			newWords := make([]Value, len(words) * len(a))
			for i := range words {
				for j := range a {
					newWords[i*len(a) + j] = words[i].Caret(ev, a[j])
				}
			}
			words = newWords
		}
	}
	return words
}

func (ev *Evaluator) evalTermList(ln *parse.ListNode) []Value {
	ev.push(ln)
	defer ev.pop()

	words := make([]Value, 0, len(ln.Nodes))
	for _, n := range ln.Nodes {
		a := ev.evalTerm(n.(*parse.ListNode))
		words = append(words, a...)
	}
	return words
}

func (ev *Evaluator) asSingleScalar(vs []Value, n parse.Node, what string) *Scalar {
	ev.push(n)
	defer ev.pop()

	if len(vs) != 1 {
		ev.errorf("Expect exactly one word for %s, got %d", what, len(vs))
	}
	v, ok := vs[0].(*Scalar)
	if !ok {
		ev.errorf("Expect scalar for %s, got %s", what, vs[0])
	}
	return v
}

func (ev *Evaluator) evalTermSingleScalar(n *parse.ListNode, what string) *Scalar {
	return ev.asSingleScalar(ev.evalTerm(n), n, what)
}

// XXX Failure of one pipeline will abort the whole chunk.
func (ev *Evaluator) evalChunk(ch *parse.ListNode) {
	for _, n := range ch.Nodes {
		updates := ev.evalPipeline(n.(*parse.ListNode))
		for i, update := range updates {
			for up := range update {
				switch up.Msg {
				case "0", "":
				default:
					// XXX Update of commands in subevaluators should not be printed.
					fmt.Printf("Command #%d update: %s\n", i, up.Msg)
				}
			}
		}
	}
}
