package eval

import (
	"fmt"
	"io"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
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
	// Pragmas tied to scopes.
	pragmas []*scopePragma
	// Destination of warning messages. This is currently only used for
	// deprecation messages.
	warn io.Writer
	// Deprecation registry.
	deprecations deprecationRegistry
	// Information about the source.
	srcMeta parse.Source
}

type scopePragma struct {
	unknownCommandIsExternal bool
}

func compile(b, g *staticNs, tree parse.Tree, w io.Writer) (op nsOp, err error) {
	g = g.clone()
	cp := &compiler{
		b, []*staticNs{g}, []*staticUpNs{new(staticUpNs)},
		[]*scopePragma{{unknownCommandIsExternal: true}},
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

// Prepares the local namespace, and returns the namespace and a function for
// executing the inner effectOp. Mutates fm.local.
func (op nsOp) prepare(fm *Frame) (*Ns, func() Exception) {
	if len(op.template.infos) > len(fm.local.infos) {
		n := len(op.template.infos)
		newLocal := &Ns{make([]vars.Var, n), op.template.infos}
		copy(newLocal.slots, fm.local.slots)
		for i := len(fm.local.infos); i < n; i++ {
			// TODO: Take readOnly into account too
			newLocal.slots[i] = MakeVarFromName(newLocal.infos[i].name)
		}
		fm.local = newLocal
	} else {
		// If no new variable has been created, there might still be some
		// existing variables deleted.
		fm.local = &Ns{fm.local.slots, op.template.infos}
	}
	return fm.local, func() Exception { return op.inner.exec(fm) }
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

func (cp *compiler) currentPragma() *scopePragma {
	return cp.pragmas[len(cp.pragmas)-1]
}

func (cp *compiler) pushScope() (*staticNs, *staticUpNs) {
	sc := new(staticNs)
	up := new(staticUpNs)
	cp.scopes = append(cp.scopes, sc)
	cp.captures = append(cp.captures, up)
	currentPragmaCopy := *cp.currentPragma()
	cp.pragmas = append(cp.pragmas, &currentPragmaCopy)
	return sc, up
}

func (cp *compiler) popScope() {
	cp.scopes[len(cp.scopes)-1] = nil
	cp.scopes = cp.scopes[:len(cp.scopes)-1]
	cp.captures[len(cp.captures)-1] = nil
	cp.captures = cp.captures[:len(cp.captures)-1]
	cp.pragmas[len(cp.pragmas)-1] = nil
	cp.pragmas = cp.pragmas[:len(cp.pragmas)-1]
}

func (cp *compiler) checkDeprecatedBuiltin(name string, r diag.Ranger) {
	msg := ""
	minLevel := 18
	// There is no builtin deprecated for 0.18.x yet. But this function is
	// only called for builtins that actually exist, so no harm checking for a
	// non-existent command here.
	switch name {
	case "deprecated-builtin~":
		msg = `the "deprecated-builtin" command is deprecated; use "new-builtin" instead`
	default:
		return
	}
	cp.deprecate(r, msg, minLevel)
}

func (cp *compiler) deprecate(r diag.Ranger, msg string, minLevel int) {
	if cp.warn == nil || r == nil {
		return
	}
	dep := deprecation{cp.srcMeta.Name, r.Range(), msg}
	if prog.DeprecationLevel >= minLevel && cp.deprecations.register(dep) {
		err := diag.Error{
			Type: "deprecation", Message: msg,
			Context: diag.Context{
				Name: cp.srcMeta.Name, Source: cp.srcMeta.Code, Ranging: r.Range()}}
		fmt.Fprintln(cp.warn, err.Show(""))
	}
}
