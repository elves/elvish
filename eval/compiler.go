package eval

//go:generate ./boilerplate.py

import (
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

type (
	scope map[string]bool
	// Op is a compiled operation.
	Op func(*EvalCtx)
	// ValuesOp is a compiled Value-generating operation.
	ValuesOp func(*EvalCtx) []Value
	// VariableOp is a compiled Variable-generating operation.
	VariableOp func(*EvalCtx) Variable
)

// compiler maintains the set of states needed when compiling a single source
// file.
type compiler struct {
	// Used in error messages.
	name, source string
	// Lexical scopes.
	scopes []scope
	// Variables captured from outer scopes.
	capture scope
	// Stored error.
	error error
}

func compile(name, source string, sc scope, n *parse.Chunk) (op Op, err error) {
	cp := &compiler{name, source, []scope{sc}, scope{}, nil}
	defer util.Catch(&err)
	return cp.chunk(n), nil
}

func (cp *compiler) errorf(p int, format string, args ...interface{}) {
	throw(util.NewContextualError(cp.name, "syntax error", cp.source, p, format, args...))
}

func (cp *compiler) thisScope() scope {
	return cp.scopes[len(cp.scopes)-1]
}

func (cp *compiler) pushScope() scope {
	sc := scope{}
	cp.scopes = append(cp.scopes, sc)
	return sc
}

func (cp *compiler) popScope() {
	cp.scopes[len(cp.scopes)-1] = nil
	cp.scopes = cp.scopes[:len(cp.scopes)-1]
}

func (cp *compiler) registerVariableGet(qname string) bool {
	_, ns, name := parseVariable(qname)
	if ns != "" && ns != "local" && ns != "up" {
		// Variable in another mod, do nothing
		return true
	}
	// Find in local scope
	if ns == "" || ns == "local" {
		if cp.thisScope()[name] {
			return true
		} else if ns == "local" {
			return false
		}
	}
	// Find in upper scopes
	for i := len(cp.scopes) - 2; i >= 0; i-- {
		if cp.scopes[i][name] {
			// Existing name: record capture and return.
			cp.capture[name] = true
			return true
		}
	}
	return false
}

func (cp *compiler) registerVariableSet(qname string) bool {
	_, ns, name := parseVariable(qname)
	switch ns {
	case "local":
		cp.thisScope()[name] = true
		return true
	case "up":
		for i := len(cp.scopes) - 2; i >= 0; i-- {
			if cp.scopes[i][name] {
				// Existing name: record capture and return.
				cp.capture[name] = true
				return true
			}
		}
		return false
	case "":
		if cp.thisScope()[name] {
			// A name on current scope. Do nothing.
			return true
		}
		// Walk up the upper scopes
		for i := len(cp.scopes) - 2; i >= 0; i-- {
			if cp.scopes[i][name] {
				// Existing name. Do nothing
				cp.capture[name] = true
				return true
			}
		}
		// New name. Register on this scope!
		cp.thisScope()[name] = true
		return true
	default:
		// Variable in another mod, do nothing
		return true
	}
}
