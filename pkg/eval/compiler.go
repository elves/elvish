package eval

import (
	"fmt"
	"io"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
)

type deprecatedWhen struct {
	minLevel int
	msg      string
}

var deprecatedVars = map[string]deprecatedWhen{
	"builtin:-source~":       {15, `the "builtin:source" command is deprecated; use "eval" instead`},
	"builtin:ord~":           {15, `the "builtin:ord" command is deprecated; use "str:to-codepoints" instead`},
	"builtin:chr~":           {15, `the "builtin:chr" command is deprecated; use "str:from-codepoints" instead`},
	"builtin:has-prefix~":    {15, `the "builtin:has-prefix" command is deprecated; use "str:has-prefix" instead`},
	"builtin:has-suffix~":    {15, `the "builtin:has-suffix" command is deprecated; use "str:has-suffix" instead`},
	"builtin:esleep~":        {15, `the "builtin:esleep" command is deprecated; use "sleep" instead`},
	"builtin:eval-symlinks~": {15, `the "builtin:eval-symlinks" command is deprecated; use "path:eval-symlinks" instead`},
	"builtin:path-abs~":      {15, `the "builtin:path-abs" command is deprecated; use "path:abs" instead`},
	"builtin:path-base~":     {15, `the "builtin:path-base" command is deprecated; use "path:base" instead`},
	"builtin:path-clean~":    {15, `the "builtin:path-clean" command is deprecated; use "path:clean" instEAD`},
	"builtin:path-dir~":      {15, `the "builtin:path-dir" command is deprecated; use "path:dir" instead`},
	"builtin:path-ext~":      {15, `the "builtin:path-ext" command is deprecated; use "path:ext" instead`},
	"builtin:-is-dir~":       {15, `the "builtin:-is-dir" command is deprecated; use "path:is-dir" instead`},
}

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

// Prepares the local namespace, and returns the namespace and a function for
// executing the inner effectOp. Mutates fm.local.
func (op nsOp) prepare(fm *Frame) (*Ns, func() Exception) {
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

func (cp *compiler) checkDeprecatedVar(name string, r diag.Ranger) {
	if dep, ok := deprecatedVars[name]; ok {
		cp.deprecate(r, dep.msg, dep.minLevel)
	}
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
