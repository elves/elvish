package eval

//go:generate ./boilerplate.py

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/diag"
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
	return Op{cp.chunkOp(n), src}, nil
}

func (cp *compiler) compiling(n parse.Node) {
	cp.begin, cp.end = n.Range().From, n.Range().To
}

func (cp *compiler) errorpf(begin, end int, format string, args ...interface{}) {
	util.Throw(NewCompilationError(fmt.Sprintf(format, args...),
		*diag.NewSourceRange(cp.srcMeta.describePath(), cp.srcMeta.code, begin, end)))
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

func (cp *compiler) registerVariableSet(qname string) bool {
	return cp.registerVariableAccess(qname, true)
}

func (cp *compiler) registerVariableGet(qname string) bool {
	return cp.registerVariableAccess(qname, false)
}

func (cp *compiler) registerVariableAccess(qname string, set bool) bool {
	readLocal := func(name string) bool { return cp.thisScope().has(name) }

	readUpvalue := func(name string) bool {
		for i := len(cp.scopes) - 2; i >= 0; i-- {
			if cp.scopes[i].has(name) {
				// Existing name: record capture and return.
				cp.capture.set(name)
				return true
			}
		}
		return false
	}

	readBuiltin := func(name string) bool { return cp.builtin.has(name) }

	readNonPseudo := func(name string) bool {
		return readLocal(name) || readUpvalue(name) || readBuiltin(name)
	}

	createLocal := func(name string) bool {
		if set && name != "" && !strings.ContainsRune(name[:len(name)-1], ':') {
			cp.thisScope().set(name)
			return true
		}
		return false
	}

	ns, name := SplitQNameNsFirst(qname) // ns = "", name = "ns:"
	name1 := name                        // name1 = "ns:"
	if name != "" && strings.ContainsRune(name[:len(name)-1], ':') {
		name1, _ = SplitQNameNsFirst(name)
	}

	// This switch mirrors the structure of that from (*Frame).ResoleVar.
	switch ns {
	case "E:":
		return true
	case "e:":
		return !set && strings.HasSuffix(name, FnSuffix)
	case "local:":
		return readLocal(name1) || createLocal(name)
	case "up:":
		return readUpvalue(name1)
	case "builtin:":
		return readBuiltin(name1)
	case "", ":":
		return readNonPseudo(name1) || createLocal(name)
	default:
		return readNonPseudo(ns)
	}
}
