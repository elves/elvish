package eval

//go:generate ./boilerplate.py

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/parse"
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
	// Information about the source.
	srcMeta *Source
}

func compile(b, g staticNs, n *parse.Chunk, src *Source) (op Op, err error) {
	cp := &compiler{b, []staticNs{g}, make(staticNs), src}
	defer func() {
		r := recover()
		if r == nil {
			return
		} else if e, ok := GetCompilationError(r); ok {
			// Save the compilation error and stop the panic.
			err = e
		} else {
			// Resume the panic; it is not supposed to be handled here.
			panic(r)
		}
	}()
	return Op{cp.chunkOp(n), src}, nil
}

func (cp *compiler) errorpf(r diag.Ranger, format string, args ...interface{}) {
	// The panic is caught by the recover in compile above.
	panic(NewCompilationError(fmt.Sprintf(format, args...),
		*diag.NewContext(cp.srcMeta.Name, cp.srcMeta.Code, r.Range().From, r.Range().To)))
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
