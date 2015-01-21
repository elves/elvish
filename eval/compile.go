package eval

import (
	"fmt"
	"os"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// Compiler compiles an Elvish AST into an Op.
type Compiler struct {
	compilerEphemeral
}

// compilerEphemeral wraps the ephemeral parts of a Compiler, namely the parts
// only valid through one startCompile-stopCompile cycle.
type compilerEphemeral struct {
	name, text string
	scopes     []map[string]Type
	enclosed   map[string]Type
}

// NewCompiler returns a new compiler.
func NewCompiler() *Compiler {
	return &Compiler{}
}

func (cp *Compiler) startCompile(name, text string, scope map[string]Type) {
	cp.compilerEphemeral = compilerEphemeral{
		name, text, []map[string]Type{scope}, make(map[string]Type),
	}
}

func (cp *Compiler) stopCompile() {
	cp.compilerEphemeral = compilerEphemeral{}
}

// Compile compiles a ChunkNode into an Op, with the knowledge of current
// scope. The supplied name and text are used in diagnostic messages.
func (cp *Compiler) Compile(name, text string, n *parse.ChunkNode, scope map[string]Type) (op Op, err error) {
	cp.startCompile(name, text, scope)
	defer cp.stopCompile()
	defer util.Recover(&err)
	return cp.compileChunk(n), nil
}

func (cp *Compiler) pushScope() {
	cp.scopes = append(cp.scopes, make(map[string]Type))
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
	util.Panic(util.NewContextualError(cp.name, "compiling error", cp.text, int(p), format, args...))
}

// compileChunk compiles a ChunkNode into an Op.
func (cp *Compiler) compileChunk(cn *parse.ChunkNode) Op {
	ops := make([]valuesOp, len(cn.Nodes))
	for i, pn := range cn.Nodes {
		ops[i] = cp.compilePipeline(pn)
	}
	return combineChunk(ops)
}

// compileClosure compiles a ClosureNode into a valuesOp along with its capture
// and the external stream types it expects.
func (cp *Compiler) compileClosure(cn *parse.ClosureNode) (valuesOp, map[string]Type) {
	ops := make([]valuesOp, len(cn.Chunk.Nodes))
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
		cp.pushVar(name, AnyType{})
	}

	for i, pn := range cn.Chunk.Nodes {
		ops[i] = cp.compilePipeline(pn)
	}

	enclosed := cp.enclosed
	cp.enclosed = make(map[string]Type)
	cp.popScope()

	return combineClosure(argNames, ops, enclosed), enclosed
}

// compilePipeline compiles a PipelineNode into a valuesOp along with the
// external stream types it expects.
func (cp *Compiler) compilePipeline(pn *parse.PipelineNode) valuesOp {
	ops := make([]stateUpdatesOp, len(pn.Nodes))

	for i, fn := range pn.Nodes {
		ops[i] = cp.compileForm(fn)
	}
	return combinePipeline(ops, pn.Pos)
}

// mustResolveVar calls ResolveVar and calls errorf if the variable is
// nonexistent.
func (cp *Compiler) mustResolveVar(name string, p parse.Pos) Type {
	if t := cp.ResolveVar(name); t != nil {
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

// ResolveVar returns the type of a variable with supplied name, found in
// current or upper scopes. If such a variable is nonexistent, a nil is
// returned. When the value to resolve is not on the current scope, it is added
// to cp.enclosed.
func (cp *Compiler) ResolveVar(name string) Type {
	if t := cp.resolveVarOnThisScope(name); t != nil {
		return t
	}
	for i := len(cp.scopes) - 2; i >= 0; i-- {
		if t := cp.scopes[i][name]; t != nil {
			cp.enclosed[name] = t
			return t
		}
	}
	return nil
}

func resolveBuiltinSpecial(cmd *parse.CompoundNode) *builtinSpecial {
	if len(cmd.Nodes) == 1 {
		sn := cmd.Nodes[0]
		if sn.Right == nil {
			pn := sn.Left
			if pn.Typ == parse.StringPrimary {
				name := pn.Node.(*parse.StringNode).Text
				if bi, ok := builtinSpecials[name]; ok {
					return &bi
				}
			}
		}
	}
	return nil
}

// compileForm compiles a FormNode into a stateUpdatesOp.
func (cp *Compiler) compileForm(fn *parse.FormNode) stateUpdatesOp {
	bi := resolveBuiltinSpecial(fn.Command)
	ports := cp.compileRedirs(fn.Redirs)

	if bi != nil {
		specialOp := bi.compile(cp, fn)
		return combineSpecialForm(specialOp, ports, fn.Pos)
	} else {
		cmdOp := cp.compileCompound(fn.Command)
		argsOp := cp.compileSpaced(fn.Args)
		return combineNonSpecialForm(cmdOp, argsOp, ports, fn.Pos)
	}
}

// compileRedirs compiles a slice of Redir's into a slice of portOp's. The
// resulting slice is indexed by the original fd.
func (cp *Compiler) compileRedirs(rs []parse.Redir) []portOp {
	var nports uintptr
	for _, rd := range rs {
		if nports < rd.Fd()+1 {
			nports = rd.Fd() + 1
		}
	}

	ports := make([]portOp, nports)
	for _, rd := range rs {
		ports[rd.Fd()] = cp.compileRedir(rd)
	}

	return ports
}

// compileRedir compiles a Redir into a portOp.
func (cp *Compiler) compileRedir(r parse.Redir) portOp {
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
		fnameOp := cp.compileCompound(r.Filename)
		return func(ev *Evaluator) *port {
			fname := string(*ev.mustSingleString(
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

// compileCompounds compiles a slice of CompoundNode's into a valuesOp. It can
// be also used to compile a SpacedNode.
func (cp *Compiler) compileCompounds(tns []*parse.CompoundNode) valuesOp {
	ops := make([]valuesOp, len(tns))
	for i, tn := range tns {
		ops[i] = cp.compileCompound(tn)
	}
	return combineSpaced(ops)
}

// compileSpaced compiles a SpacedNode into a valuesOp.
func (cp *Compiler) compileSpaced(ln *parse.SpacedNode) valuesOp {
	return cp.compileCompounds(ln.Nodes)
}

// compileCompound compiles a CompoundNode into a valuesOp.
func (cp *Compiler) compileCompound(tn *parse.CompoundNode) valuesOp {
	var op valuesOp
	if len(tn.Nodes) == 1 {
		op = cp.compileSubscript(tn.Nodes[0])
	} else {
		ops := make([]valuesOp, len(tn.Nodes))
		for i, fn := range tn.Nodes {
			ops[i] = cp.compileSubscript(fn)
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

// compileSubscript compiles a SubscriptNode into a valuesOp.
func (cp *Compiler) compileSubscript(sn *parse.SubscriptNode) valuesOp {
	if sn.Right == nil {
		return cp.compilePrimary(sn.Left)
	}
	left := cp.compilePrimary(sn.Left)
	right := cp.compileCompound(sn.Right)
	return combineSubscript(cp, left, right, sn.Left.Pos, sn.Right.Pos)
}

// compilePrimary compiles a PrimaryNode into a valuesOp.
func (cp *Compiler) compilePrimary(fn *parse.PrimaryNode) valuesOp {
	switch fn.Typ {
	case parse.StringPrimary:
		text := fn.Node.(*parse.StringNode).Text
		return makeString(text)
	case parse.VariablePrimary:
		name := fn.Node.(*parse.StringNode).Text
		return makeVar(cp, name, fn.Pos)
	case parse.TablePrimary:
		table := fn.Node.(*parse.TableNode)
		list := cp.compileCompounds(table.List)
		keys := make([]valuesOp, len(table.Dict))
		values := make([]valuesOp, len(table.Dict))
		for i, tp := range table.Dict {
			keys[i] = cp.compileCompound(tp.Key)
			values[i] = cp.compileCompound(tp.Value)
		}
		return combineTable(list, keys, values, fn.Pos)
	case parse.ClosurePrimary:
		op, enclosed := cp.compileClosure(fn.Node.(*parse.ClosureNode))
		// Added variables enclosed on inner closures to cp.enclosed
		for name, typ := range enclosed {
			if cp.resolveVarOnThisScope(name) == nil {
				cp.enclosed[name] = typ
			}
		}
		return op
	case parse.ListPrimary:
		return cp.compileSpaced(fn.Node.(*parse.SpacedNode))
	case parse.ChanCapturePrimary:
		op := cp.compilePipeline(fn.Node.(*parse.PipelineNode))
		return combineChanCapture(op)
	case parse.StatusCapturePrimary:
		op := cp.compilePipeline(fn.Node.(*parse.PipelineNode))
		return op
	default:
		panic(fmt.Sprintln("bad PrimaryNode type", fn.Typ))
	}
}
