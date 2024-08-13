package eval

import (
	"fmt"
	"io"
	"strings"

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
	// Names of internal modules.
	modules []string
	// Destination of warning messages. This is currently only used for
	// deprecation messages.
	warn io.Writer
	// Deprecation registry.
	deprecations deprecationRegistry
	// Information about the source.
	src parse.Source
	// Compilation errors.
	errors []*CompilationError
	// Suggested code to fix potential issues found during compilation.
	autofixes []string
}

type scopePragma struct {
	unknownCommandIsExternal bool
}

func compile(b, g *staticNs, modules []string, tree parse.Tree, w io.Writer) (nsOp, []string, error) {
	g = g.clone()
	cp := &compiler{
		b, []*staticNs{g}, []*staticUpNs{new(staticUpNs)},
		[]*scopePragma{{unknownCommandIsExternal: true}},
		modules,
		w, newDeprecationRegistry(), tree.Source, nil, nil}
	chunkOp := cp.chunkOp(tree.Root)
	return nsOp{chunkOp, g}, cp.autofixes, diag.PackErrors(cp.errors)
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

type CompilationError = diag.Error[CompilationErrorTag]

// CompilationErrorTag parameterizes [diag.Error] to define [CompilationError].
type CompilationErrorTag struct{}

func (CompilationErrorTag) ErrorTag() string { return "compilation error" }

// Reports a compilation error.
func (cp *compiler) errorpf(r diag.Ranger, format string, args ...any) {
	cp.errorpfInner(r, fmt.Sprintf(format, args...), false)
}

// Reports a compilation error, and mark it as partial iff the end of r happens
// to coincide with the end of the source code.
func (cp *compiler) errorpfPartial(r diag.Ranger, format string, args ...any) {
	cp.errorpfInner(r, fmt.Sprintf(format, args...), r.Range().To == len(cp.src.Code))
}

func (cp *compiler) errorpfInner(r diag.Ranger, msg string, partial bool) {
	cp.errors = append(cp.errors, &CompilationError{
		Message: msg,
		Context: *diag.NewContext(cp.src.Name, cp.src.Code, r),
		// TODO: This criteria is too strict and only captures a small subset of
		// partial compilation errors.
		Partial: partial,
	})
}

// UnpackCompilationErrors returns the constituent compilation errors if the
// given error contains one or more compilation errors. Otherwise it returns
// nil.
func UnpackCompilationErrors(e error) []*CompilationError {
	if errs := diag.UnpackErrors[CompilationErrorTag](e); len(errs) > 0 {
		return errs
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
	minLevel := 22
	switch name {
	// We don't have any deprecated builtins targeted for 0.22 yet, but keep
	// this code here so that the code doesn't get stale. This function is only
	// called for symbols that actually resolve to builtins, so having a fake
	// one here is harmless.
	case "foo~":
		msg = `the "foo" command is deprecated; use "bar" instead`
	default:
		return
	}
	cp.deprecate(r, msg, minLevel)
}

type deprecationTag struct{}

func (deprecationTag) ErrorTag() string { return "deprecation" }

func (cp *compiler) deprecate(r diag.Ranger, msg string, minLevel int) {
	if cp.warn == nil || r == nil {
		return
	}
	dep := deprecation{cp.src.Name, r.Range(), msg}
	if prog.DeprecationLevel >= minLevel && cp.deprecations.register(dep) {
		err := diag.Error[deprecationTag]{
			Message: msg,
			Context: *diag.NewContext(cp.src.Name, cp.src.Code, r.Range())}
		fmt.Fprintln(cp.warn, err.Show(""))
	}
}

// Given a variable that doesn't resolve, add any applicable autofixes.
func (cp *compiler) autofixUnresolvedVar(qname string) {
	if len(cp.modules) == 0 {
		return
	}
	first, _ := SplitQName(qname)
	mod := strings.TrimSuffix(first, ":")
	if mod != first && sliceContains(cp.modules, mod) {
		cp.autofixes = append(cp.autofixes, "use "+mod)
	}
}
