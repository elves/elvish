package eval

import (
	"fmt"
	"io"
	"strings"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/prog"
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
	// New variables created in a lexical scope.
	newLocals []string
	// Destination of warning messages. This is currently only used for
	// deprecation messages.
	warn io.Writer
	// Deprecation registry.
	deprecations deprecationRegistry
	// Information about the source.
	srcMeta parse.Source
}

func compile(b, g staticNs, tree parse.Tree, w io.Writer) (op Op, err error) {
	cp := &compiler{
		b, []staticNs{g}, make(staticNs), nil,
		w, newDeprecationRegistry(), tree.Source}
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
	savedLocals := cp.pushNewLocals()
	chunkOp := cp.chunkOp(tree.Root)
	scopeOp := wrapScopeOp(chunkOp, cp.newLocals)
	cp.newLocals = savedLocals

	return Op{scopeOp, tree.Source}, nil
}

func (cp *compiler) pushNewLocals() []string {
	saved := cp.newLocals
	cp.newLocals = nil
	return saved
}

func (cp *compiler) errorpf(r diag.Ranger, format string, args ...interface{}) {
	// The panic is caught by the recover in compile above.
	panic(NewCompilationError(fmt.Sprintf(format, args...),
		*diag.NewContext(cp.srcMeta.Name, cp.srcMeta.Code, r)))
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
	return cp.registerVariableAccess(qname, true, nil)
}

func (cp *compiler) registerVariableGet(qname string, r diag.Ranger) bool {
	return cp.registerVariableAccess(qname, false, r)
}

func (cp *compiler) registerVariableAccess(qname string, set bool, r diag.Ranger) bool {
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

	readBuiltin := func(name string) bool {
		cp.checkDeprecatedBuiltin(name, r)
		return cp.builtin.has(name)
	}

	readNonPseudo := func(name string) bool {
		return readLocal(name) || readUpvalue(name) || readBuiltin(name)
	}

	createLocal := func(name string) bool {
		if set && name != "" && !strings.ContainsRune(name[:len(name)-1], ':') {
			cp.thisScope().set(name)
			cp.newLocals = append(cp.newLocals, name)
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

func (cp *compiler) checkDeprecatedBuiltin(name string, r diag.Ranger) {
	if cp.warn == nil || r == nil {
		return
	}
	msg := ""
	switch name {
	case "explode~":
		msg = `the "explode" command is deprecated; use "all" instead`
	case "join~":
		msg = `the "joins" command is deprecated; use "str:join" instead`
	case "splits~":
		msg = `the "splits" command is deprecated; use "str:split" instead`
	case "replaces~":
		msg = `the "replaces" command is deprecated; use "str:replace" instead`
	case "-time~":
		msg = `the "-time" command is deprecated; use "time" instead`
	default:
		return
	}
	dep := deprecation{cp.srcMeta.Name, r.Range(), msg}
	if prog.ShowDeprecations && cp.deprecations.register(dep) {
		err := diag.Error{
			Type: "deprecation", Message: msg,
			Context: diag.Context{
				Name: cp.srcMeta.Name, Source: cp.srcMeta.Code, Ranging: r.Range()}}
		fmt.Fprintln(cp.warn, err.Show(""))
	}
}
