package eval

import (
	"fmt"
	"io"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/parse"
)

// compiler maintains the set of states needed when compiling a single source
// file.
type compiler struct {
	// Builtin namespace.
	builtin *staticNs
	// Lexical namespaces.
	scopes []*staticNs
	// Sources of captured variables.
	captures []*staticUpNs
	// Destination of warning messages. This is currently only used for
	// deprecation messages.
	warn io.Writer
	// Deprecation registry.
	deprecations deprecationRegistry
	// Information about the source.
	srcMeta parse.Source
}

type capture struct {
	name string
	// If true, the captured variable is from the immediate outer level scope,
	// i.e. the local scope the lambda is evaluated in. Otherwise the captured
	// variable is from a more outer level, i.e. the upvalue scope the lambda is
	// evaluated in.
	local bool
	// Index to the captured variable.
	index int
}

func compile(b, g *staticNs, tree parse.Tree, w io.Writer) (op nsOp, err error) {
	g = g.clone()
	cp := &compiler{
		b, []*staticNs{g}, []*staticUpNs{new(staticUpNs)},
		w, newDeprecationRegistry(), tree.Source}
	defer func() {
		r := recover()
		if r == nil {
			return
		} else if e := GetCompilationError(r); e != nil {
			// Save the compilation error and stop the panic.
			err = e
		} else {
			// Resume the panic; it is not supposed to be handled here.
			panic(r)
		}
	}()
	chunkOp := cp.chunkOp(tree.Root)
	return nsOp{chunkOp, g}, nil
}

type nsOp struct {
	inner    effectOp
	template *staticNs
}

// Prepares a new local namespace before executing the inner effectOp. Replaces
// fm.local.
func (op nsOp) exec(fm *Frame) Exception {
	op.prepareNs(fm)
	return op.inner.exec(fm)
}

func (op nsOp) prepareNs(fm *Frame) {
	if len(op.template.names) > len(fm.local.names) {
		n := len(op.template.names)
		newLocal := &Ns{make([]vars.Var, n), op.template.names, op.template.deleted}
		copy(newLocal.slots, fm.local.slots)
		for i := len(fm.local.names); i < n; i++ {
			newLocal.slots[i] = MakeVarFromName(newLocal.names[i])
		}
		fm.local = newLocal
	} else {
		// If no new has been created, there might still be some existing
		// variables deleted.
		fm.local = &Ns{fm.local.slots, fm.local.names, op.template.deleted}
	}

}

const compilationErrorType = "compilation error"

func (cp *compiler) errorpf(r diag.Ranger, format string, args ...interface{}) {
	// The panic is caught by the recover in compile above.
	panic(&diag.Error{
		Type:    compilationErrorType,
		Message: fmt.Sprintf(format, args...),
		Context: *diag.NewContext(cp.srcMeta.Name, cp.srcMeta.Code, r)})
}

// GetCompilationError returns a *diag.Error if the given value is a compilation
// error. Otherwise it returns nil.
func GetCompilationError(e interface{}) *diag.Error {
	if e, ok := e.(*diag.Error); ok && e.Type == compilationErrorType {
		return e
	}
	return nil
}
func (cp *compiler) thisScope() *staticNs {
	return cp.scopes[len(cp.scopes)-1]
}

func (cp *compiler) pushScope() (*staticNs, *staticUpNs) {
	sc := new(staticNs)
	up := new(staticUpNs)
	cp.scopes = append(cp.scopes, sc)
	cp.captures = append(cp.captures, up)
	return sc, up
}

func (cp *compiler) popScope() {
	cp.scopes[len(cp.scopes)-1] = nil
	cp.scopes = cp.scopes[:len(cp.scopes)-1]
	cp.captures[len(cp.captures)-1] = nil
	cp.captures = cp.captures[:len(cp.captures)-1]
}
