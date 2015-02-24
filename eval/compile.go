package eval

import (
	"fmt"
	"os"
	"strings"

	"github.com/elves/elvish/errutil"
	"github.com/elves/elvish/parse"
)

// staticNS is the static type information of a namespace.
type staticNS map[string]Type

// Compiler compiles an Elvish AST into an Op.
type Compiler struct {
	builtin staticNS
	scopes  []staticNS
	up      staticNS
	mod     map[string]staticNS
	dataDir string
	compilerEphemeral
}

// compilerEphemeral wraps the ephemeral parts of a Compiler, namely the parts
// only valid through one startCompile-stopCompile cycle.
type compilerEphemeral struct {
	name, text, dir string
}

// NewCompiler returns a new compiler.
func NewCompiler(bi staticNS, dataDir string) *Compiler {
	return &Compiler{
		bi, []staticNS{staticNS{}}, staticNS{}, map[string]staticNS{}, dataDir,
		compilerEphemeral{},
	}
}

func (cp *Compiler) startCompile(name, text, dir string) {
	cp.compilerEphemeral = compilerEphemeral{name, text, dir}
}

func (cp *Compiler) stopCompile() {
	cp.compilerEphemeral = compilerEphemeral{}
}

// Compile compiles a ChunkNode into an Op. The supplied name and text are used
// in diagnostic messages.
func (cp *Compiler) Compile(name, text, dir string, n *parse.Chunk) (op Op, err error) {
	cp.startCompile(name, text, dir)
	defer cp.stopCompile()
	defer errutil.Catch(&err)
	return cp.chunk(n), nil
}

func (cp *Compiler) pushScope() {
	cp.scopes = append(cp.scopes, staticNS{})
}

func (cp *Compiler) popScope() {
	cp.scopes[len(cp.scopes)-1] = nil
	cp.scopes = cp.scopes[:len(cp.scopes)-1]
}

func (cp *Compiler) pushVar(name string, t Type) {
	cp.scopes[len(cp.scopes)-1][name] = t
}

func (cp *Compiler) popVar(name string) {
	delete(cp.scopes[len(cp.scopes)-1], name)
}

func (cp *Compiler) errorf(p parse.Pos, format string, args ...interface{}) {
	errutil.Throw(errutil.NewContextualError(cp.name, "compiling error", cp.text, int(p), format, args...))
}

// chunk compiles a ChunkNode into an Op.
func (cp *Compiler) chunk(cn *parse.Chunk) Op {
	ops := make([]valuesOp, len(cn.Nodes))
	for i, pn := range cn.Nodes {
		ops[i] = cp.pipeline(pn)
	}
	return combineChunk(ops)
}

// closure compiles a ClosureNode into a valuesOp.
func (cp *Compiler) closure(cn *parse.Closure) valuesOp {
	nargs := 0
	if cn.ArgNames != nil {
		nargs = len(cn.ArgNames.Nodes)
	}
	argNames := make([]string, nargs)

	if nargs > 0 {
		// TODO Allow types for arguments. Maybe share code with the var
		// builtin.
		for i, cn := range cn.ArgNames.Nodes {
			_, name := ensureVariablePrimary(cp, cn, "expect variable")
			argNames[i] = name
		}
	}

	cp.pushScope()

	for _, name := range argNames {
		cp.pushVar(name, anyType{})
	}

	op := cp.chunk(cn.Chunk)

	up := cp.up
	cp.up = staticNS{}
	cp.popScope()

	// Added variables up on inner closures to cp.up
	for name, typ := range up {
		if cp.resolveVarOnThisScope(name) == nil {
			cp.up[name] = typ
		}
	}

	return combineClosure(argNames, op, up)
}

// pipeline compiles a PipelineNode into a valuesOp.
func (cp *Compiler) pipeline(pn *parse.Pipeline) valuesOp {
	ops := make([]stateUpdatesOp, len(pn.Nodes))

	for i, fn := range pn.Nodes {
		ops[i] = cp.form(fn)
	}
	return combinePipeline(ops, pn.Pos)
}

// mustResolveVar calls ResolveVar and calls errorf if the variable is
// nonexistent.
func (cp *Compiler) mustResolveVar(ns, name string, p parse.Pos) Type {
	if t := cp.ResolveVar(ns, name); t != nil {
		return t
	}
	cp.errorf(p, "undefined variable $%s", name)
	return nil
}

// resolveVarOnThisScope returns the type of the named variable on current
// scope. When such a variable does not exist, nil is returned.
func (cp *Compiler) resolveVarOnThisScope(name string) Type {
	return cp.scopes[len(cp.scopes)-1][name]
}

// splitQualifiedName splits a qualified variable name into two parts separated
// by a colon, the namespace and the name proper. When there is no colon, the
// namespace part is empty.
func splitQualifiedName(qname string) (string, string) {
	i := strings.IndexRune(qname, ':')
	if i == -1 {
		return "", qname
	}
	return qname[:i], qname[i+1:]
}

// ResolveVar returns the type of a variable with supplied name and on the
// supplied namespace. If such a variable is nonexistent, a nil is returned.
// When the value to resolve is on an outer current scope, it is added to
// cp.up.
func (cp *Compiler) ResolveVar(ns, name string) Type {
	if ns == "env" {
		return stringType{}
	}
	if mod, ok := cp.mod[ns]; ok {
		return mod[name]
	}

	may := func(n string) bool {
		return ns == "" || ns == n
	}
	if may("local") {
		if t := cp.resolveVarOnThisScope(name); t != nil {
			return t
		}
	}
	if may("up") {
		for i := len(cp.scopes) - 2; i >= 0; i-- {
			if t, ok := cp.scopes[i][name]; ok {
				cp.up[name] = t
				return t
			}
		}
	}
	if may("builtin") {
		if t, ok := cp.builtin[name]; ok {
			return t
		}
	}
	return nil
}

func resolveBuiltinSpecial(cmd *parse.Compound) *builtinSpecial {
	if len(cmd.Nodes) == 1 {
		sn := cmd.Nodes[0]
		if sn.Right == nil {
			pn := sn.Left
			if pn.Typ == parse.StringPrimary {
				name := pn.Node.(*parse.String).Text
				if bi, ok := builtinSpecials[name]; ok {
					return &bi
				}
			}
		}
	}
	return nil
}

// form compiles a FormNode into a stateUpdatesOp.
func (cp *Compiler) form(fn *parse.Form) stateUpdatesOp {
	bi := resolveBuiltinSpecial(fn.Command)
	ports := cp.redirs(fn.Redirs)

	if bi != nil {
		specialOp := bi.compile(cp, fn)
		return combineSpecialForm(specialOp, ports, fn.Pos)
	}
	cmdOp := cp.compound(fn.Command)
	argsOp := cp.spaced(fn.Args)
	return combineNonSpecialForm(cmdOp, argsOp, ports, fn.Pos)
}

// redirs compiles a slice of Redir's into a slice of portOp's. The resulting
// slice is indexed by the original fd.
func (cp *Compiler) redirs(rs []parse.Redir) []portOp {
	var nports uintptr
	for _, rd := range rs {
		if nports < rd.Fd()+1 {
			nports = rd.Fd() + 1
		}
	}

	ports := make([]portOp, nports)
	for _, rd := range rs {
		ports[rd.Fd()] = cp.redir(rd)
	}

	return ports
}

// redir compiles a Redir into a portOp.
func (cp *Compiler) redir(r parse.Redir) portOp {
	switch r := r.(type) {
	case *parse.CloseRedir:
		return func(ev *Evaluator) *port {
			return &port{}
		}
	case *parse.FdRedir:
		oldFd := int(r.OldFd)
		return func(ev *Evaluator) *port {
			// Copied ports have shouldClose unmarked to avoid double close on
			// channels
			p := *ev.port(oldFd)
			p.closeF = false
			p.closeCh = false
			return &p
		}
	case *parse.FilenameRedir:
		fnameOp := cp.compound(r.Filename)
		return func(ev *Evaluator) *port {
			fname := string(ev.mustSingleString(
				fnameOp.f(ev), "filename", r.Filename.Pos))
			// TODO haz hardcoded permbits now
			f, e := os.OpenFile(fname, r.Flag, 0644)
			if e != nil {
				ev.errorf(r.Pos, "failed to open file %q: %s", fname[0], e)
			}
			return &port{
				f: f, ch: make(chan Value), closeF: true, closeCh: true,
			}
		}
	default:
		panic("bad Redir type")
	}
}

// compounds compiles a slice of CompoundNode's into a valuesOp. It can
// be also used to compile a SpacedNode.
func (cp *Compiler) compounds(tns []*parse.Compound) valuesOp {
	ops := make([]valuesOp, len(tns))
	for i, tn := range tns {
		ops[i] = cp.compound(tn)
	}
	return combineSpaced(ops)
}

// spaced compiles a SpacedNode into a valuesOp.
func (cp *Compiler) spaced(ln *parse.Spaced) valuesOp {
	return cp.compounds(ln.Nodes)
}

// compound compiles a CompoundNode into a valuesOp.
func (cp *Compiler) compound(tn *parse.Compound) valuesOp {
	var op valuesOp
	if len(tn.Nodes) == 1 {
		op = cp.subscript(tn.Nodes[0])
	} else {
		ops := make([]valuesOp, len(tn.Nodes))
		for i, fn := range tn.Nodes {
			ops[i] = cp.subscript(fn)
		}
		op = combineCompound(ops)
	}
	if tn.Sigil == parse.NoSigil {
		return op
	}
	cmdOp := makeString(string(tn.Sigil))
	fop := combineNonSpecialForm(cmdOp, op, nil, tn.Pos)
	pop := combinePipeline([]stateUpdatesOp{fop}, tn.Pos)
	return combineChanCapture(pop)
}

// subscript compiles a SubscriptNode into a valuesOp.
func (cp *Compiler) subscript(sn *parse.Subscript) valuesOp {
	if sn.Right == nil {
		return cp.primary(sn.Left)
	}
	left := cp.primary(sn.Left)
	right := cp.compound(sn.Right)
	return combineSubscript(cp, left, right, sn.Left.Pos, sn.Right.Pos)
}

// primary compiles a PrimaryNode into a valuesOp.
func (cp *Compiler) primary(fn *parse.Primary) valuesOp {
	switch fn.Typ {
	case parse.StringPrimary:
		text := fn.Node.(*parse.String).Text
		return makeString(text)
	case parse.VariablePrimary:
		name := fn.Node.(*parse.String).Text
		return makeVar(cp, name, fn.Pos)
	case parse.TablePrimary:
		table := fn.Node.(*parse.Table)
		list := cp.compounds(table.List)
		keys := make([]valuesOp, len(table.Dict))
		values := make([]valuesOp, len(table.Dict))
		for i, tp := range table.Dict {
			keys[i] = cp.compound(tp.Key)
			values[i] = cp.compound(tp.Value)
		}
		return combineTable(list, keys, values, fn.Pos)
	case parse.ClosurePrimary:
		return cp.closure(fn.Node.(*parse.Closure))
	case parse.ListPrimary:
		return cp.spaced(fn.Node.(*parse.Spaced))
	case parse.ChanCapturePrimary:
		op := cp.pipeline(fn.Node.(*parse.Pipeline))
		return combineChanCapture(op)
	case parse.StatusCapturePrimary:
		op := cp.pipeline(fn.Node.(*parse.Pipeline))
		return op
	default:
		panic(fmt.Sprintln("bad PrimaryNode type", fn.Typ))
	}
}
