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

func (cp *Compiler) errorf(n parse.Node, format string, args ...interface{}) {
	util.Panic(util.NewContextualError(cp.name, cp.text, int(n.Position()), format, args...))
}

func (cp *Compiler) compileChunk(cn *parse.ChunkNode) Op {
	ops := make([]valuesOp, len(cn.Nodes))
	for i, pn := range cn.Nodes {
		ops[i], _ = cp.compilePipeline(pn)
	}
	return combineChunk(ops)
}

func (cp *Compiler) compileClosure(cn *parse.ClosureNode) (valuesOp, *closureAnnotation) {
	ops := make([]valuesOp, len(cn.Chunk.Nodes))

	cp.pushScope()
	annotation := &closureAnnotation{}
	cn.Annotation = annotation

	bounds := [2]StreamType{}
	for i, pn := range cn.Chunk.Nodes {
		var a *pipelineAnnotation
		ops[i], a = cp.compilePipeline(pn)

		var ok bool
		bounds[0], ok = bounds[0].commonType(a.bounds[0])
		if !ok {
			cp.errorf(pn, "Pipeline input stream incompatible with previous ones")
		}
		bounds[1], ok = bounds[1].commonType(a.bounds[1])
		if !ok {
			cp.errorf(pn, "Pipeline output stream incompatible with previous ones")
		}
	}
	annotation.bounds = bounds

	annotation.enclosed = cp.enclosed
	cp.enclosed = make(map[string]Type)
	cp.popScope()

	return combineClosure(ops, annotation), annotation
}

func (cp *Compiler) compilePipeline(pn *parse.PipelineNode) (valuesOp, *pipelineAnnotation) {
	ops := make([]stateUpdatesOp, len(pn.Nodes))
	annotation := &pipelineAnnotation{}

	var lastOutput StreamType
	for i, fn := range pn.Nodes {
		var a *formAnnotation
		ops[i], a = cp.compileForm(fn)
		input := a.streamTypes[0]
		if i == 0 {
			annotation.bounds[0] = input
		} else {
			internal, ok := lastOutput.commonType(input)
			if !ok {
				cp.errorf(fn, "Form input type %v insatisfiable - previous form output is type %v", input, lastOutput)
			}
			annotation.internals = append(annotation.internals, internal)
		}
		lastOutput = a.streamTypes[1]
	}
	annotation.bounds[1] = lastOutput
	return combinePipeline(pn, ops, annotation), annotation
}

func (cp *Compiler) resolveVar(name string, n *parse.FactorNode) Type {
	if t := cp.tryResolveVar(name); t != nil {
		return t
	}
	cp.errorf(n, "undefined variable $%q", name)
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

func (cp *Compiler) compileForm(fn *parse.FormNode) (stateUpdatesOp, *formAnnotation) {
	// TODO(xiaq): Allow more interesting terms to be used as commands
	msg := "command must be a string or closure"
	if len(fn.Command.Nodes) != 1 {
		cp.errorf(fn.Command, msg)
	}
	command := fn.Command.Nodes[0]
	cmdOp := cp.compileFactor(command)

	annotation := &formAnnotation{}
	switch command.Typ {
	case parse.StringFactor:
		cp.resolveCommand(command.Node.(*parse.StringNode).Text, annotation)
	case parse.ClosureFactor:
		ca := command.Node.(*parse.ClosureNode).Annotation.(*closureAnnotation)
		annotation.commandType = commandClosure
		annotation.streamTypes = ca.bounds
	default:
		cp.errorf(fn.Command, msg)
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
					cp.errorf(rd, "fd redir on channel port")
				}
			case *parse.FilenameRedir:
				if annotation.streamTypes[fd] == chanStream {
					cp.errorf(rd, "filename redir on channel port")
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
		tlist = cp.compileTermList(fn.Args)
	}
	return combineForm(fn, cmdOp, tlist, ports, annotation), annotation
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
		fnameOp := cp.compileTerm(r.Filename)
		return func(ev *Evaluator) *port {
			fname := string(*ev.asSingleString(
				r.Filename, fnameOp(ev), "filename"))
			// TODO haz hardcoded permbits now
			f, e := os.OpenFile(fname, r.Flag, 0644)
			if e != nil {
				ev.errorfNode(r, "failed to open file %q: %s", fname[0], e)
			}
			return &port{f: f, shouldClose: true}
		}
	default:
		panic("bad Redir type")
	}
}

func (cp *Compiler) compileTerms(tns []*parse.TermNode) valuesOp {
	ops := make([]valuesOp, len(tns))
	for i, tn := range tns {
		ops[i] = cp.compileTerm(tn)
	}
	return combineTermList(ops)
}

func (cp *Compiler) compileTermList(ln *parse.TermListNode) valuesOp {
	return cp.compileTerms(ln.Nodes)
}

func (cp *Compiler) compileTerm(tn *parse.TermNode) valuesOp {
	ops := make([]valuesOp, len(tn.Nodes))
	for i, fn := range tn.Nodes {
		ops[i] = cp.compileFactor(fn)
	}
	return combineTerm(ops)
}

func (cp *Compiler) compileFactor(fn *parse.FactorNode) valuesOp {
	switch fn.Typ {
	case parse.StringFactor:
		text := fn.Node.(*parse.StringNode).Text
		return makeString(text)
	case parse.VariableFactor:
		name := fn.Node.(*parse.StringNode).Text
		cp.resolveVar(name, fn)
		return makeVar(name)
	case parse.TableFactor:
		table := fn.Node.(*parse.TableNode)
		list := cp.compileTerms(table.List)
		keys := make([]valuesOp, len(table.Dict))
		values := make([]valuesOp, len(table.Dict))
		for i, tp := range table.Dict {
			keys[i] = cp.compileTerm(tp.Key)
			values[i] = cp.compileTerm(tp.Value)
		}
		return combineTable(fn, list, keys, values)
	case parse.ClosureFactor:
		op, ca := cp.compileClosure(fn.Node.(*parse.ClosureNode))
		for name, typ := range ca.enclosed {
			if !cp.hasVarOnThisScope(name) {
				cp.enclosed[name] = typ
			}
		}
		return op
	case parse.ListFactor:
		return cp.compileTermList(fn.Node.(*parse.TermListNode))
	case parse.OutputCaptureFactor:
		op, a := cp.compilePipeline(fn.Node.(*parse.PipelineNode))
		return combineOutputCapture(op, a)
	case parse.StatusCaptureFactor:
		op, _ := cp.compilePipeline(fn.Node.(*parse.PipelineNode))
		return op
	default:
		panic(fmt.Sprintln("bad FactorNode type", fn.Typ))
	}
}
