// Package eval handles evaluation of nodes and consists the runtime of the
// shell.
package eval

import (
	"os"
	"strings"
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

func evalFactor(n *parse.FactorNode) ([]string, error) {
	switch n := n.Node.(type) {
	case *parse.StringNode:
		return []string{n.Text}, nil
	case *parse.ListNode:
		return evalTermList(n)
	default:
		panic("bad node type")
	}
}

func evalTerm(n_ parse.Node) ([]string, error) {
	n := n_.(*parse.ListNode)
	words := make([]string, 0, len(n.Nodes))
	words = append(words, "")
	for _, m := range n.Nodes {
		a, e := evalFactor(m.(*parse.FactorNode))
		if e != nil {
			return nil, e
		}
		if len(a) == 1 {
			for i := range words {
				words[i] += a[0]
			}
		} else {
			// Do a Cartesian product
			newWords := make([]string, len(words) * len(a))
			for i := range words {
				for j := range a {
					newWords[i*len(a) + j] = words[i] + a[j]
				}
			}
			words = newWords
		}
	}
	return words, nil
}

func evalTermList(ln *parse.ListNode) ([]string, error) {
	words := make([]string, 0, len(ln.Nodes))
	for _, n := range ln.Nodes {
		a, e := evalTerm(n)
		if e != nil {
			return nil, e
		}
		words = append(words, a...)
	}
	return words, nil
}
