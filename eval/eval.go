// Package eval handles evaluation of nodes and consists the runtime of the
// shell.
package eval

import (
	"fmt"
	"strings"
	"syscall"
	"strconv"
	"../parse"
)

type Evaluator struct {
	env map[string]string
	searchPaths []string
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
func (ev *Evaluator) resolveVar(name string) (Value, error) {
	if name == "!pid" {
		return NewScalar(strconv.Itoa(syscall.Getpid())), nil
	}
	val, ok := ev.env[name]
	if !ok {
		return nil, fmt.Errorf("Variable not found: %s", name)
	}
	return NewScalar(val), nil
}

func (ev *Evaluator) evalTable(tn *parse.TableNode) (*Table, error) {
	// Evaluate list part.
	t := NewTable()
	for _, term := range tn.List {
		vs, err := ev.evalTerm(term)
		if err != nil {
			return nil, err
		}
		t.append(vs...)
	}
	for _, pair := range tn.Dict {
		ks, err := ev.evalTerm(pair.Key)
		if err != nil {
			return nil, err
		}
		vs, err := ev.evalTerm(pair.Value)
		if err != nil {
			return nil, err
		}
		if len(ks) != len(vs) {
			return nil, fmt.Errorf("Number of keys doesn't match number of values: %d vs. %d", len(ks), len(vs))
		}
		for i, k := range ks {
			t.dict[k] = vs[i]
		}
	}
	return t, nil
}

func (ev *Evaluator) evalFactor(n *parse.FactorNode) ([]Value, error) {
	var words []Value
	var err error

	switch n := n.Node.(type) {
	case *parse.StringNode:
		words = []Value{NewScalar(n.Text)}
	case *parse.ListNode:
		words, err = ev.evalTermList(n)
		if err != nil {
			return nil, err
		}
	case *parse.TableNode:
		word, err := ev.evalTable(n)
		if err != nil {
			return nil, err
		}
		words = []Value{word}
	default:
		panic("bad node type")
	}

	for dollar := n.Dollar; dollar > 0; dollar-- {
		if len(words) != 1 {
			return nil, fmt.Errorf("Only a single value may be dollared")
		}
		if _, ok := words[0].(*Scalar); !ok {
			return nil, fmt.Errorf("Only scalar may be dollared (for now)")
		}
		words[0], err = ev.resolveVar(words[0].(*Scalar).str)
		if err != nil {
			return nil, err
		}
	}

	return words, nil
}

func (ev *Evaluator) evalTerm(n *parse.ListNode) ([]Value, error) {
	if len(n.Nodes) == 1 {
		// Only one factor.
		return ev.evalFactor(n.Nodes[0].(*parse.FactorNode))
	}
	// More than one factor.
	// The result is always a scalar list.
	words := make([]Value, 1, len(n.Nodes))
	words[0] = NewScalar("")
	for _, m := range n.Nodes {
		a, e := ev.evalFactor(m.(*parse.FactorNode))
		if e != nil {
			return nil, e
		}
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
	return words, nil
}

func (ev *Evaluator) evalTermList(ln *parse.ListNode) ([]Value, error) {
	words := make([]Value, 0, len(ln.Nodes))
	for _, n := range ln.Nodes {
		a, e := ev.evalTerm(n.(*parse.ListNode))
		if e != nil {
			return nil, e
		}
		words = append(words, a...)
	}
	return words, nil
}
