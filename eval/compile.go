package eval

import (
	"fmt"
	"os"

	"github.com/xiaq/elvish/parse"
	"github.com/xiaq/elvish/util"
)

// Compiler compiles an Elvish AST into an Op.
type Compiler struct {
	compilerEphemeral
}

// compilerEphemeral wraps the ephemeral parts of a Compiler.
type compilerEphemeral struct {
	name, text string
	scopes     []map[string]Type
	enclosed   map[string]Type
}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (cp *Compiler) startCompile(name, text string, scope map[string]Type) {
	cp.compilerEphemeral = compilerEphemeral{
		name, text, []map[string]Type{scope}, make(map[string]Type),
	}
}

func (cp *Compiler) Compile(name, text string, n *parse.ChunkNode, scope map[string]Type) (op Op, err error) {
	cp.startCompile(name, text, scope)
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

func (cp *Compiler) hasVarOnThisScope(name string) bool {
	_, ok := cp.scopes[len(cp.scopes)-1][name]
	return ok
}

func (cp *Compiler) errorf(p parse.Pos, format string, args ...interface{}) {
	util.Panic(util.NewContextualError(cp.name, cp.text, int(p), format, args...))
}

func (cp *Compiler) compileChunk(cn *parse.ChunkNode) Op {
	ops := make([]valuesOp, len(cn.Nodes))
	for i, pn := range cn.Nodes {
		ops[i], _ = cp.compilePipeline(pn)
	}
	return combineChunk(ops)
}

func (cp *Compiler) compileClosure(cn *parse.ClosureNode) (valuesOp, map[string]Type, [2]StreamType) {
	ops := make([]valuesOp, len(cn.Chunk.Nodes))

	cp.pushScope()

	bounds := [2]StreamType{}
	for i, pn := range cn.Chunk.Nodes {
		var b [2]StreamType
		ops[i], b = cp.compilePipeline(pn)

		var ok bool
		bounds[0], ok = bounds[0].commonType(b[0])
		if !ok {
			cp.errorf(pn.Pos, "Pipeline input stream incompatible with previous ones")
		}
		bounds[1], ok = bounds[1].commonType(b[1])
		if !ok {
			cp.errorf(pn.Pos, "Pipeline output stream incompatible with previous ones")
		}
	}

	enclosed := cp.enclosed
	cp.enclosed = make(map[string]Type)
	cp.popScope()

	return combineClosure(ops, enclosed, bounds), enclosed, bounds
}

func (cp *Compiler) compilePipeline(pn *parse.PipelineNode) (valuesOp, [2]StreamType) {
	ops := make([]stateUpdatesOp, len(pn.Nodes))
	var bounds [2]StreamType
	internals := make([]StreamType, len(pn.Nodes)-1)

	var lastOutput StreamType
	for i, fn := range pn.Nodes {
		var b [2]StreamType
		ops[i], b = cp.compileForm(fn)
		input := b[0]
		if i == 0 {
			bounds[0] = input
		} else {
			internal, ok := lastOutput.commonType(input)
			if !ok {
				cp.errorf(fn.Pos, "Form input type %v insatisfiable - previous form output is type %v", input, lastOutput)
			}
			internals[i-1] = internal
		}
		lastOutput = b[1]
	}
	bounds[1] = lastOutput
	return combinePipeline(ops, bounds, internals, pn.Pos), bounds
}

func (cp *Compiler) resolveVar(name string, p parse.Pos) Type {
	if t := cp.tryResolveVar(name); t != nil {
		return t
	}
	cp.errorf(p, "undefined variable $%s", name)
	return nil
}

func (cp *Compiler) tryResolveVar(name string) Type {
	thisScope := len(cp.scopes) - 1
	for i := thisScope; i >= 0; i-- {
		if t := cp.scopes[i][name]; t != nil {
			if i < thisScope {
				cp.enclosed[name] = t
			}
			return t
		}
	}
	return nil
}

func (cp *Compiler) resolveCommand(name string, fa *formAnnotation) {
	if ct, ok := cp.tryResolveVar("fn-" + name).(*ClosureType); ok {
		// Defined function
		fa.commandType = commandDefinedFunction
		fa.streamTypes = ct.Bounds
	} else if bi, ok := builtinSpecials[name]; ok {
		// Builtin special
		fa.commandType = commandBuiltinSpecial
		fa.streamTypes = bi.streamTypes
		fa.builtinSpecial = &bi
	} else if bi, ok := builtinFuncs[name]; ok {
		// Builtin func
		fa.commandType = commandBuiltinFunction
		fa.streamTypes = bi.streamTypes
		fa.builtinFunc = &bi
	} else {
		// External command
		fa.commandType = commandExternal
		fa.streamTypes = [2]StreamType{fdStream, fdStream}
	}
}

func (cp *Compiler) compileForm(fn *parse.FormNode) (stateUpdatesOp, [2]StreamType) {
	// TODO(xiaq): Allow more interesting compound expressions to be used as
	// commands
	msg := "command must be a string or closure"
	if len(fn.Command.Nodes) != 1 || fn.Command.Nodes[0].Right != nil {
		cp.errorf(fn.Command.Pos, msg)
	}
	command := fn.Command.Nodes[0].Left
	cmdOp, pbounds := cp.compilePrimary(command)

	annotation := &formAnnotation{}
	switch command.Typ {
	case parse.StringPrimary:
		cp.resolveCommand(command.Node.(*parse.StringNode).Text, annotation)
	case parse.ClosurePrimary:
		annotation.commandType = commandClosure
		annotation.streamTypes = *pbounds
	default:
		cp.errorf(fn.Command.Pos, msg)
	}

	var nports uintptr
	for _, rd := range fn.Redirs {
		if nports < rd.Fd()+1 {
			nports = rd.Fd() + 1
		}
	}

	ports := make([]portOp, nports)
	for _, rd := range fn.Redirs {
		fd := rd.Fd()
		if fd < 2 {
			switch rd := rd.(type) {
			case *parse.FdRedir:
				if annotation.streamTypes[fd] == chanStream {
					cp.errorf(rd.Pos, "fd redir on channel port")
				}
			case *parse.FilenameRedir:
				if annotation.streamTypes[fd] == chanStream {
					cp.errorf(rd.Pos, "filename redir on channel port")
				}
			}
			annotation.streamTypes[fd] = unusedStream
		}
		ports[fd] = cp.compileRedir(rd)
	}

	var tlist valuesOp
	if annotation.commandType == commandBuiltinSpecial {
		annotation.specialOp = annotation.builtinSpecial.compile(cp, fn)
	} else {
		tlist = cp.compileSpaced(fn.Args)
	}
	return combineForm(cmdOp, tlist, ports, annotation, fn.Pos), annotation.streamTypes
}

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
			p.shouldClose = false
			return &p
		}
	case *parse.FilenameRedir:
		fnameOp := cp.compileCompound(r.Filename)
		return func(ev *Evaluator) *port {
			fname := string(*ev.asSingleString(
				fnameOp.f(ev), "filename", r.Filename.Pos))
			// TODO haz hardcoded permbits now
			f, e := os.OpenFile(fname, r.Flag, 0644)
			if e != nil {
				ev.errorfPos(r.Pos, "failed to open file %q: %s", fname[0], e)
			}
			return &port{f: f, shouldClose: true}
		}
	default:
		panic("bad Redir type")
	}
}

func (cp *Compiler) compileCompounds(tns []*parse.CompoundNode) valuesOp {
	ops := make([]valuesOp, len(tns))
	for i, tn := range tns {
		ops[i] = cp.compileCompound(tn)
	}
	return combineSpaced(ops)
}

func (cp *Compiler) compileSpaced(ln *parse.SpacedNode) valuesOp {
	return cp.compileCompounds(ln.Nodes)
}

func (cp *Compiler) compileCompound(tn *parse.CompoundNode) valuesOp {
	ops := make([]valuesOp, len(tn.Nodes))
	for i, fn := range tn.Nodes {
		ops[i], _ = cp.compileSubscript(fn)
	}
	return combineCompound(ops)
}

func (cp *Compiler) compileSubscript(sn *parse.SubscriptNode) (valuesOp, *[2]StreamType) {
	if sn.Right == nil {
		return cp.compilePrimary(sn.Left)
	}
	left, _ := cp.compilePrimary(sn.Left)
	right := cp.compileCompound(sn.Right)
	return combineSubscript(cp, left, right, sn.Left.Pos, sn.Right.Pos), nil
}

func (cp *Compiler) compilePrimary(fn *parse.PrimaryNode) (valuesOp, *[2]StreamType) {
	switch fn.Typ {
	case parse.StringPrimary:
		text := fn.Node.(*parse.StringNode).Text
		return makeString(text), nil
	case parse.VariablePrimary:
		name := fn.Node.(*parse.StringNode).Text
		return makeVar(cp, name, fn.Pos), nil
	case parse.TablePrimary:
		table := fn.Node.(*parse.TableNode)
		list := cp.compileCompounds(table.List)
		keys := make([]valuesOp, len(table.Dict))
		values := make([]valuesOp, len(table.Dict))
		for i, tp := range table.Dict {
			keys[i] = cp.compileCompound(tp.Key)
			values[i] = cp.compileCompound(tp.Value)
		}
		return combineTable(list, keys, values, fn.Pos), nil
	case parse.ClosurePrimary:
		op, enclosed, bounds := cp.compileClosure(fn.Node.(*parse.ClosureNode))
		for name, typ := range enclosed {
			if !cp.hasVarOnThisScope(name) {
				cp.enclosed[name] = typ
			}
		}
		return op, &bounds
	case parse.ListPrimary:
		return cp.compileSpaced(fn.Node.(*parse.SpacedNode)), nil
	case parse.OutputCapturePrimary:
		op, b := cp.compilePipeline(fn.Node.(*parse.PipelineNode))
		return combineOutputCapture(op, b), nil
	case parse.StatusCapturePrimary:
		op, _ := cp.compilePipeline(fn.Node.(*parse.PipelineNode))
		return op, nil
	default:
		panic(fmt.Sprintln("bad PrimaryNode type", fn.Typ))
	}
}
