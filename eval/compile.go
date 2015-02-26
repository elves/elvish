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
	global  staticNS
	builtin staticNS
	// scopes  []staticNS
	mod     map[string]staticNS
	dataDir string
}

// compileCtx maintains a Compiler along with mutable states. After creation a
// compilerCtx is never modified (but its fields may be), and new instances are
// created when needed.
type compileCtx struct {
	*Compiler
	name, text, dir string

	scopes []staticNS
	up     staticNS
}

func newCompileCtx(cp *Compiler, name, text, dir string) *compileCtx {
	return &compileCtx{
		cp,
		name, text, dir,
		[]staticNS{cp.global}, staticNS{},
	}
}

// NewCompiler returns a new compiler.
func NewCompiler(bi staticNS, dataDir string) *Compiler {
	return &Compiler{staticNS{}, bi, map[string]staticNS{}, dataDir}
}

// Compile compiles a ChunkNode into an Op. The supplied name and text are used
// in diagnostic messages.
func (cp *Compiler) Compile(name, text, dir string, n *parse.Chunk) (Op, error) {
	cc := newCompileCtx(cp, name, text, dir)
	return cc.compile(n)
}

func (cc *compileCtx) compile(n *parse.Chunk) (op Op, err error) {
	defer errutil.Catch(&err)
	return cc.chunk(n), nil
}

func (cc *compileCtx) pushScope() {
	cc.scopes = append(cc.scopes, staticNS{})
}

func (cc *compileCtx) popScope() {
	cc.scopes[len(cc.scopes)-1] = nil
	cc.scopes = cc.scopes[:len(cc.scopes)-1]
}

func (cc *compileCtx) pushVar(name string, t Type) {
	cc.scopes[len(cc.scopes)-1][name] = t
}

func (cc *compileCtx) popVar(name string) {
	delete(cc.scopes[len(cc.scopes)-1], name)
}

func (cc *compileCtx) errorf(p parse.Pos, format string, args ...interface{}) {
	errutil.Throw(errutil.NewContextualError(cc.name, "compiling error", cc.text, int(p), format, args...))
}

// chunk compiles a ChunkNode into an Op.
func (cc *compileCtx) chunk(cn *parse.Chunk) Op {
	ops := make([]valuesOp, len(cn.Nodes))
	for i, pn := range cn.Nodes {
		ops[i] = cc.pipeline(pn)
	}
	return combineChunk(ops)
}

// closure compiles a ClosureNode into a valuesOp.
func (cc *compileCtx) closure(cn *parse.Closure) valuesOp {
	nargs := 0
	if cn.ArgNames != nil {
		nargs = len(cn.ArgNames.Nodes)
	}
	argNames := make([]string, nargs)

	if nargs > 0 {
		// TODO Allow types for arguments. Maybe share code with the var
		// builtin.
		for i, cn := range cn.ArgNames.Nodes {
			_, name := ensureVariablePrimary(cc, cn, "expect variable")
			argNames[i] = name
		}
	}

	cc.pushScope()

	for _, name := range argNames {
		cc.pushVar(name, anyType{})
	}

	op := cc.chunk(cn.Chunk)

	up := cc.up
	cc.up = staticNS{}
	cc.popScope()

	// Added variables up on inner closures to cc.up
	for name, typ := range up {
		if cc.resolveVarOnThisScope(name) == nil {
			cc.up[name] = typ
		}
	}

	return combineClosure(argNames, op, up)
}

// pipeline compiles a PipelineNode into a valuesOp.
func (cc *compileCtx) pipeline(pn *parse.Pipeline) valuesOp {
	ops := make([]stateUpdatesOp, len(pn.Nodes))

	for i, fn := range pn.Nodes {
		ops[i] = cc.form(fn)
	}
	return combinePipeline(ops, pn.Pos)
}

// mustResolveVar calls ResolveVar and calls errorf if the variable is
// nonexistent.
func (cc *compileCtx) mustResolveVar(ns, name string, p parse.Pos) Type {
	if t := cc.ResolveVar(ns, name); t != nil {
		return t
	}
	cc.errorf(p, "undefined variable $%s", name)
	return nil
}

// resolveVarOnThisScope returns the type of the named variable on current
// scope. When such a variable does not exist, nil is returned.
func (cc *compileCtx) resolveVarOnThisScope(name string) Type {
	return cc.scopes[len(cc.scopes)-1][name]
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
// cc.up.
func (cc *compileCtx) ResolveVar(ns, name string) Type {
	if ns == "env" {
		return stringType{}
	}
	if mod, ok := cc.mod[ns]; ok {
		return mod[name]
	}

	may := func(n string) bool {
		return ns == "" || ns == n
	}
	if may("local") {
		if t := cc.resolveVarOnThisScope(name); t != nil {
			return t
		}
	}
	if may("up") {
		for i := len(cc.scopes) - 2; i >= 0; i-- {
			if t, ok := cc.scopes[i][name]; ok {
				cc.up[name] = t
				return t
			}
		}
	}
	if may("builtin") {
		if t, ok := cc.builtin[name]; ok {
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
func (cc *compileCtx) form(fn *parse.Form) stateUpdatesOp {
	bi := resolveBuiltinSpecial(fn.Command)
	ports := cc.redirs(fn.Redirs)

	if bi != nil {
		specialOp := bi.compile(cc, fn)
		return combineSpecialForm(specialOp, ports, fn.Pos)
	}
	cmdOp := cc.compound(fn.Command)
	argsOp := cc.spaced(fn.Args)
	return combineNonSpecialForm(cmdOp, argsOp, ports, fn.Pos)
}

// redirs compiles a slice of Redir's into a slice of portOp's. The resulting
// slice is indexed by the original fd.
func (cc *compileCtx) redirs(rs []parse.Redir) []portOp {
	var nports uintptr
	for _, rd := range rs {
		if nports < rd.Fd()+1 {
			nports = rd.Fd() + 1
		}
	}

	ports := make([]portOp, nports)
	for _, rd := range rs {
		ports[rd.Fd()] = cc.redir(rd)
	}

	return ports
}

const defaultFileRedirPerm = 0644

// redir compiles a Redir into a portOp.
func (cc *compileCtx) redir(r parse.Redir) portOp {
	switch r := r.(type) {
	case *parse.CloseRedir:
		return func(ec *evalCtx) *port {
			return &port{}
		}
	case *parse.FdRedir:
		oldFd := int(r.OldFd)
		return func(ec *evalCtx) *port {
			// Copied ports have shouldClose unmarked to avoid double close on
			// channels
			p := *ec.port(oldFd)
			p.closeF = false
			p.closeCh = false
			return &p
		}
	case *parse.FilenameRedir:
		fnameOp := cc.compound(r.Filename)
		return func(ec *evalCtx) *port {
			fname := string(ec.mustSingleString(
				fnameOp.f(ec), "filename", r.Filename.Pos))
			f, e := os.OpenFile(fname, r.Flag, defaultFileRedirPerm)
			if e != nil {
				ec.errorf(r.Pos, "failed to open file %q: %s", fname[0], e)
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
func (cc *compileCtx) compounds(tns []*parse.Compound) valuesOp {
	ops := make([]valuesOp, len(tns))
	for i, tn := range tns {
		ops[i] = cc.compound(tn)
	}
	return combineSpaced(ops)
}

// spaced compiles a SpacedNode into a valuesOp.
func (cc *compileCtx) spaced(ln *parse.Spaced) valuesOp {
	return cc.compounds(ln.Nodes)
}

// compound compiles a CompoundNode into a valuesOp.
func (cc *compileCtx) compound(tn *parse.Compound) valuesOp {
	var op valuesOp
	if len(tn.Nodes) == 1 {
		op = cc.subscript(tn.Nodes[0])
	} else {
		ops := make([]valuesOp, len(tn.Nodes))
		for i, fn := range tn.Nodes {
			ops[i] = cc.subscript(fn)
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
func (cc *compileCtx) subscript(sn *parse.Subscript) valuesOp {
	if sn.Right == nil {
		return cc.primary(sn.Left)
	}
	left := cc.primary(sn.Left)
	right := cc.compound(sn.Right)
	return combineSubscript(cc, left, right, sn.Left.Pos, sn.Right.Pos)
}

// primary compiles a PrimaryNode into a valuesOp.
func (cc *compileCtx) primary(fn *parse.Primary) valuesOp {
	switch fn.Typ {
	case parse.StringPrimary:
		text := fn.Node.(*parse.String).Text
		return makeString(text)
	case parse.VariablePrimary:
		name := fn.Node.(*parse.String).Text
		return makeVar(cc, name, fn.Pos)
	case parse.TablePrimary:
		table := fn.Node.(*parse.Table)
		list := cc.compounds(table.List)
		keys := make([]valuesOp, len(table.Dict))
		values := make([]valuesOp, len(table.Dict))
		for i, tp := range table.Dict {
			keys[i] = cc.compound(tp.Key)
			values[i] = cc.compound(tp.Value)
		}
		return combineTable(list, keys, values, fn.Pos)
	case parse.ClosurePrimary:
		return cc.closure(fn.Node.(*parse.Closure))
	case parse.ListPrimary:
		return cc.spaced(fn.Node.(*parse.Spaced))
	case parse.ChanCapturePrimary:
		op := cc.pipeline(fn.Node.(*parse.Pipeline))
		return combineChanCapture(op)
	case parse.StatusCapturePrimary:
		op := cc.pipeline(fn.Node.(*parse.Pipeline))
		return op
	default:
		panic(fmt.Sprintln("bad PrimaryNode type", fn.Typ))
	}
}
