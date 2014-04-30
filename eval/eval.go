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
	Compiler    *Compiler
	name, text  string
	scope       map[string]*Value
	env         *Env
	searchPaths []string
	in, out     *port
	statusCb    func([]Value)
	nodes       []parse.Node // A stack that keeps track of nodes being evaluated.
}

func statusOk(vs []Value) bool {
	for _, v := range vs {
		v, ok := v.(*String)
		if !ok {
			return false
		}
		switch string(*v) {
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
	g := map[string]*Value{
		"env": valuePtr(env), "pid": valuePtr(pid),
	}
	ev := &Evaluator{
		Compiler: &Compiler{},
		scope:    g, env: env,
		in: &port{f: os.Stdin}, out: &port{f: os.Stdout},
	}
	ev.statusCb = func(vs []Value) {
		if statusOk(vs) {
			return
		}
		fmt.Print("Status: ")
		for i, v := range vs {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(v.Repr(ev))
		}
		fmt.Println()
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

func (ev *Evaluator) MakeCompilerScope() map[string]Type {
	scope := make(map[string]Type)
	for name, value := range ev.scope {
		scope[name] = (*value).Type()
	}
	return scope
}

// Eval evaluates a chunk node n. The name and text of it is used for
// diagnostic messages.
func (ev *Evaluator) Eval(name, text string, n *parse.ChunkNode) error {
	op, err := ev.Compiler.Compile(name, text, n, ev.MakeCompilerScope())
	if err != nil {
		return err
	}
	return ev.eval(name, text, op)
}

func (ev *Evaluator) eval(name, text string, op Op) (err error) {
	if op == nil {
		return nil
	}
	defer util.Recover(&err)
	defer ev.stopEval()
	ev.name = name
	ev.text = text
	op(ev)
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

func (ev *Evaluator) asSingleString(n parse.Node, vs []Value, what string) *String {
	if len(vs) != 1 {
		ev.errorfNode(n, "Expect exactly one word for %s, got %d", what, len(vs))
	}
	v, ok := vs[0].(*String)
	if !ok {
		ev.errorfNode(n, "Expect string for %s, got %s", what, vs[0])
	}
	return v
}
