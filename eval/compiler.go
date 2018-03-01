package eval

//go:generate ./boilerplate.py

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// compiler maintains the set of states needed when compiling a single source
// file.
type compiler struct {
	// Builtin namespace.
	builtin staticNs
	// Lexical namespaces.
	scopes []staticNs
	// Variables captured from outer scopes.
	capture staticNs
	// Position of what is being compiled.
	begin, end int
	// Information about the source.
	srcMeta *Source
}

func compile(b, g staticNs, n *parse.Chunk, src *Source) (op Op, err error) {
	cp := &compiler{b, []staticNs{g}, make(staticNs), 0, 0, src}
	defer util.Catch(&err)
	return cp.chunkOp(n), nil
}

func (cp *compiler) compiling(n parse.Node) {
	cp.begin, cp.end = n.Begin(), n.End()
}

func (cp *compiler) errorpf(begin, end int, format string, args ...interface{}) {
	throw(&CompilationError{fmt.Sprintf(format, args...),
		*util.NewSourceRange(cp.srcMeta.describePath(), cp.srcMeta.code, begin, end)})
}

func (cp *compiler) errorf(format string, args ...interface{}) {
	cp.errorpf(cp.begin, cp.end, format, args...)
}

func (cp *compiler) thisScope() staticNs {
	return cp.scopes[len(cp.scopes)-1]
}

func (cp *compiler) pushScope() staticNs {
	sc := make(staticNs)
	cp.scopes = append(cp.scopes, sc)
	return sc
}

func (cp *compiler) popScope() {
	cp.scopes[len(cp.scopes)-1] = make(staticNs)
	cp.scopes = cp.scopes[:len(cp.scopes)-1]
}

func (cp *compiler) registerVariableGetQname(qname string) bool {
	_, ns, name := ParseVariableRef(qname)
	return cp.registerVariableGet(ns, name)
}

func (cp *compiler) registerVariableGet(ns, name string) bool {
	switch ns {
	case "", "local", "up":
		// Handled below
	case "e", "E":
		return true
	default:
		return cp.registerModAccess(ns)
	}
	// Find in local scope
	if ns == "" || ns == "local" {
		if cp.thisScope().has(name) {
			return true
		}
	}
	// Find in upper scopes
	if ns == "" || ns == "up" {
		for i := len(cp.scopes) - 2; i >= 0; i-- {
			if cp.scopes[i].has(name) {
				// Existing name: record capture and return.
				cp.capture.set(name)
				return true
			}
		}
	}
	// Find in builtin scope
	if ns == "" || ns == "builtin" {
		if cp.builtin.has(name) {
			return true
		}
	}
	return false
}

func (cp *compiler) registerVariableSetQname(qname string) bool {
	_, ns, name := ParseVariableRef(qname)
	return cp.registerVariableSet(ns, name)
}

func (cp *compiler) registerVariableSet(ns, name string) bool {
	switch ns {
	case "local":
		cp.thisScope().set(name)
		return true
	case "up":
		for i := len(cp.scopes) - 2; i >= 0; i-- {
			if cp.scopes[i].has(name) {
				// Existing name: record capture and return.
				cp.capture.set(name)
				return true
			}
		}
		return false
	case "builtin":
		cp.errorf("cannot set builtin variable")
		return false
	case "":
		if cp.thisScope().has(name) {
			// A name on current scope. Do nothing.
			return true
		}
		// Walk up the upper scopes
		for i := len(cp.scopes) - 2; i >= 0; i-- {
			if cp.scopes[i].has(name) {
				// Existing name. Do nothing
				cp.capture.set(name)
				return true
			}
		}
		// New name. Register on this scope!
		cp.thisScope().set(name)
		return true
	case "e", "E":
		// Special namespaces, do nothing
		return true
	default:
		return cp.registerModAccess(ns)
	}
}

func (cp *compiler) registerModAccess(name string) bool {
	if strings.ContainsRune(name, ':') {
		name = name[:strings.IndexByte(name, ':')]
	}
	return cp.registerVariableGet("", name+NsSuffix)
}
