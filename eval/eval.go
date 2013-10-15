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

var env map[string]string
var search_paths []string

func init() {
	env = envAsMap(os.Environ())

	path_var, ok := env["PATH"]
	if ok {
		search_paths = strings.Split(path_var, ":")
		// fmt.Printf("Search paths are %v\n", search_paths)
	} else {
		search_paths = []string{"/bin"}
	}
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
func resolveVar(name string) (Value, error) {
	if name == "!pid" {
		return NewScalar(strconv.Itoa(syscall.Getpid())), nil
	}
	val, ok := env[name]
	if !ok {
		return nil, fmt.Errorf("Variable not found: %s", name)
	}
	return NewScalar(val), nil
}

func evalFactor(n *parse.FactorNode) ([]Value, error) {
	var words []Value
	var err error

	switch n := n.Node.(type) {
	case *parse.StringNode:
		words = []Value{NewScalar(n.Text)}
	case *parse.ListNode:
		words, err = evalTermList(n)
		if err != nil {
			return nil, err
		}
	default:
		panic("bad node type")
	}

	if n.Dollar {
		for i := range words {
			// XXX Assumes scalar word.
			words[i], err = resolveVar(words[i].(*Scalar).str)
			if err != nil {
				return nil, err
			}
		}
	}

	return words, nil
}

func evalTerm(n *parse.ListNode) ([]Value, error) {
	if len(n.Nodes) == 1 {
		// Only one factor.
		return evalFactor(n.Nodes[0].(*parse.FactorNode))
	}
	// More than one factor.
	// The result is always a scalar list.
	words := make([]Value, 1, len(n.Nodes))
	words[0] = NewScalar("")
	for _, m := range n.Nodes {
		a, e := evalFactor(m.(*parse.FactorNode))
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

func evalTermList(ln *parse.ListNode) ([]Value, error) {
	words := make([]Value, 0, len(ln.Nodes))
	for _, n := range ln.Nodes {
		a, e := evalTerm(n.(*parse.ListNode))
		if e != nil {
			return nil, e
		}
		words = append(words, a...)
	}
	return words, nil
}
