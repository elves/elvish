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
)

type Evaluator struct {
	env map[string]string
	searchPaths []string
	filesToClose []*os.File
}

func NewEvaluator(env []string) *Evaluator {
	ev := &Evaluator{env: envAsMap(env)}

	path, ok := ev.env["PATH"]
	if ok {
		ev.searchPaths = strings.Split(path, ":")
		// fmt.Printf("Search paths are %v\n", search_paths)
	} else {
		ev.searchPaths = []string{"/bin"}
	}

	return ev
}

// errorf stops the evaluator. Its panic is supposed to be caught by recover.
func (ev *Evaluator) errorf(n parse.Node, format string, args...interface{}) {
	panic(&Error{n, fmt.Sprintf(format, args...)})
}

// recover is the handler that turns panics into returns from top level of
// evaluation function (currently ExecPipeline).
func (ev *Evaluator) recover(perr **Error) {
	r := recover()
	if r == nil {
		return
	}
	if _, ok := r.(*Error); !ok {
		panic(r)
	}
	*perr = r.(*Error)
}

func envAsMap(env []string) (m map[string]string) {
	m = make(map[string]string)
	for _, e := range env {
		arr := strings.SplitN(e, "=", 2)
		if len(arr) == 2 {
			m[arr[0]] = arr[1]
		}
	}
	return
}

// XXX Makes up scalar values from env on the fly.
func (ev *Evaluator) resolveVar(name string, n parse.Node) Value {
	if name == "!pid" {
		return NewScalar(strconv.Itoa(syscall.Getpid()))
	}
	val, ok := ev.env[name]
	if !ok {
		ev.errorf(n, "Variable not found: %s", name)
	}
	return NewScalar(val)
}

func (ev *Evaluator) evalTable(tn *parse.TableNode) *Table {
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
			ev.errorf(tn, "Number of keys doesn't match number of values: %d vs. %d", len(ks), len(vs))
		}
		for i, k := range ks {
			t.dict[k] = vs[i]
		}
	}
	return t
}

func (ev *Evaluator) evalFactor(n *parse.FactorNode) []Value {
	var words []Value

	switch n := n.Node.(type) {
	case *parse.StringNode:
		words = []Value{NewScalar(n.Text)}
	case *parse.ListNode:
		words = ev.evalTermList(n)
	case *parse.TableNode:
		word := ev.evalTable(n)
		words = []Value{word}
	default:
		panic("bad node type")
	}

	for dollar := n.Dollar; dollar > 0; dollar-- {
		if len(words) != 1 {
			ev.errorf(n, "Only a single value may be dollared")
		}
		if _, ok := words[0].(*Scalar); !ok {
			ev.errorf(n, "Only scalar may be dollared (for now)")
		}
		words[0] = ev.resolveVar(words[0].(*Scalar).str, n)
	}

	return words
}

func (ev *Evaluator) evalTerm(n *parse.ListNode) []Value {
	if len(n.Nodes) == 1 {
		// Only one factor.
		return ev.evalFactor(n.Nodes[0].(*parse.FactorNode))
	}
	// More than one factor.
	// The result is always a scalar list.
	words := make([]Value, 1, len(n.Nodes))
	words[0] = NewScalar("")
	for _, m := range n.Nodes {
		a := ev.evalFactor(m.(*parse.FactorNode))
		if len(a) == 1 {
			for i := range words {
				words[i].(*Scalar).str += a[0].String()
			}
		} else {
			// Do a Cartesian product
			newWords := make([]Value, len(words) * len(a))
			for i := range words {
				for j := range a {
					newWords[i*len(a) + j] = NewScalar(words[i].String() + a[j].String())
				}
			}
			words = newWords
		}
	}
	return words
}

func (ev *Evaluator) evalTermList(ln *parse.ListNode) []Value {
	words := make([]Value, 0, len(ln.Nodes))
	for _, n := range ln.Nodes {
		a := ev.evalTerm(n.(*parse.ListNode))
		words = append(words, a...)
	}
	return words
}
