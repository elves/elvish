package eval

//go:generate ./boilerplate.py

import (
	"fmt"
	"strconv"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// compiler maintains the set of states needed when compiling a single source
// file.
type compiler struct {
	// Builtin scope.
	builtin staticScope
	// Lexical scopes.
	scopes []staticScope
	// Variables captured from outer scopes.
	capture staticScope
	// Position of what is being compiled.
	begin, end int
	// Information about the source.
	name, text string
}

func compile(b, g staticScope, n *parse.Chunk, name, text string) (op Op, err error) {
	cp := &compiler{b, []staticScope{g}, makeStaticScope(), 0, 0, name, text}
	defer util.Catch(&err)
	return cp.chunkOp(n), nil
}

func (cp *compiler) compiling(n parse.Node) {
	cp.begin, cp.end = n.Begin(), n.End()
}

func (cp *compiler) errorpf(begin, end int, format string, args ...interface{}) {
	throw(&CompilationError{fmt.Sprintf(format, args...),
		*util.NewSourceRange(cp.name, cp.text, begin, end, nil)})
}

func (cp *compiler) errorf(format string, args ...interface{}) {
	cp.errorpf(cp.begin, cp.end, format, args...)
}

func (cp *compiler) thisScope() staticScope {
	return cp.scopes[len(cp.scopes)-1]
}

func (cp *compiler) pushScope() staticScope {
	sc := makeStaticScope()
	cp.scopes = append(cp.scopes, sc)
	return sc
}

func (cp *compiler) popScope() {
	cp.scopes[len(cp.scopes)-1] = makeStaticScope()
	cp.scopes = cp.scopes[:len(cp.scopes)-1]
}

func (cp *compiler) registerVariableGet(qname string) bool {
	_, ns, name := ParseVariable(qname)
	switch ns {
	case "", "local", "up":
		// Handled below
	case "e", "E", "shared":
		return true
	default:
		return cp.registerModAccess(ns)
	}
	_, err := strconv.Atoi(name)
	isnum := err == nil
	// Find in local scope
	if ns == "" || ns == "local" {
		if cp.thisScope().Names[name] || isnum {
			return true
		}
	}
	// Find in upper scopes
	if ns == "" || ns == "up" {
		for i := len(cp.scopes) - 2; i >= 0; i-- {
			if cp.scopes[i].Names[name] || isnum {
				// Existing name: record capture and return.
				cp.capture.Names[name] = true
				return true
			}
		}
	}
	// Find in builtin scope
	if ns == "" || ns == "builtin" {
		_, ok := cp.builtin.Names[name]
		if ok {
			return true
		}
	}
	return false
}

func (cp *compiler) registerVariableSet(qname string) bool {
	_, ns, name := ParseVariable(qname)
	switch ns {
	case "local":
		cp.thisScope().Names[name] = true
		return true
	case "up":
		for i := len(cp.scopes) - 2; i >= 0; i-- {
			if cp.scopes[i].Names[name] {
				// Existing name: record capture and return.
				cp.capture.Names[name] = true
				return true
			}
		}
		return false
	case "builtin":
		cp.errorf("cannot set builtin variable")
		return false
	case "":
		if cp.thisScope().Names[name] {
			// A name on current scope. Do nothing.
			return true
		}
		// Walk up the upper scopes
		for i := len(cp.scopes) - 2; i >= 0; i-- {
			if cp.scopes[i].Names[name] {
				// Existing name. Do nothing
				cp.capture.Names[name] = true
				return true
			}
		}
		// New name. Register on this scope!
		cp.thisScope().Names[name] = true
		return true
	case "e", "E", "shared":
		// Special namespaces, do nothing
		return true
	default:
		return cp.registerModAccess(ns)
	}
}

func (cp *compiler) registerModAccess(name string) bool {
	if cp.thisScope().Uses[name] {
		return true
	}
	for i := len(cp.scopes) - 2; i >= 0; i-- {
		if cp.scopes[i].Uses[name] {
			cp.capture.Uses[name] = true
			return true
		}
	}
	return cp.builtin.Uses[name]
}
