package eval

import (
	"github.com/xiaq/elvish/parse"
	"github.com/xiaq/elvish/util"
)

// Checker performs static checking on an Elvish AST. It also annotates the AST
// with static information that is useful during evaluation.
type Checker struct {
	name, text string
	scopes     []map[string]bool
}

func NewChecker() *Checker {
	return &Checker{}
}

func (ch *Checker) Check(name, text string, n *parse.ChunkNode, scope map[string]bool) (err error) {
	ch.name = name
	ch.text = text
	ch.scopes = []map[string]bool{scope}

	defer util.Recover(&err)
	ch.checkChunk(n)
	return nil
}

func (ch *Checker) pushScope() {
	ch.scopes = append(ch.scopes, make(map[string]bool))
}

func (ch *Checker) popScope() {
	ch.scopes[len(ch.scopes)-1] = nil
	ch.scopes = ch.scopes[:len(ch.scopes)-1]
}

func (ch *Checker) pushVar(name string) {
	ch.scopes[len(ch.scopes)-1][name] = true
}

func (ch *Checker) errorf(n parse.Node, format string, args ...interface{}) {
	util.Panic(util.NewContextualError(ch.name, ch.text, int(n.Position()), format, args...))
}

// checkChunk checks a ChunkNode by checking all pipelines it contains.
func (ch *Checker) checkChunk(cn *parse.ChunkNode) {
	for _, pn := range cn.Nodes {
		ch.checkPipeline(pn)
	}
}

// checkClosure checks a ClosureNode by checking the chunk it contains.
// TODO(xiaq): Check that all pipelines have coherent IO ports.
func (ch *Checker) checkClosure(cn *parse.ClosureNode) {
}

// checkPipeline checks a PipelineNode by checking all forms and checking that
// all connected ports are compatible. It also annotates the node.
func (ch *Checker) checkPipeline(pn *parse.PipelineNode) {
	for _, fn := range pn.Nodes {
		ch.checkForm(fn)
	}
	annotation := &pipelineAnnotation{}
	pn.Annotation = annotation
	annotation.bounds[0] = pn.Nodes[0].Annotation.(*formAnnotation).streamTypes[0]
	annotation.bounds[1] = pn.Nodes[len(pn.Nodes)-1].Annotation.(*formAnnotation).streamTypes[1]
}

func (ch *Checker) resolveVar(name string, n *parse.FactorNode) {
	if !ch.tryResolveVar(name) {
		ch.errorf(n, "undefined variable $%q", name)
	}
}

func (ch *Checker) tryResolveVar(name string) bool {
	// XXX(xiaq): Variables in outer scopes ("enclosed variables") are resolved
	// correctly by the checker by not by the evaluator.
	for i := len(ch.scopes) - 1; i >= 0; i-- {
		if ch.scopes[i][name] {
			return true
		}
	}
	return false
}

func (ch *Checker) resolveCommand(name string, fa *formAnnotation) {
	if ch.tryResolveVar("fn-" + name) {
		// Defined function
		// XXX(xiaq): Assume fdStream IO for closures
		fa.commandType = commandDefinedFunction
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
		// Just use zero value (fdStream) for fa.streamTypes
	}
}

// checkForm checks a FormNode by resolving the command statically and checking
// all terms. Special forms are then processed case by case. It also annotates
// the node.
func (ch *Checker) checkForm(fn *parse.FormNode) {
	// TODO(xiaq): Allow more interesting terms to be used as commands
	msg := "command must be a string or closure"
	if len(fn.Command.Nodes) != 1 {
		ch.errorf(fn.Command, msg)
	}
	command := fn.Command.Nodes[0]
	annotation := &formAnnotation{}
	fn.Annotation = annotation
	switch command.Typ {
	case parse.StringFactor:
		ch.resolveCommand(command.Node.(*parse.StringNode).Text, annotation)
	case parse.ClosureFactor:
		// XXX(xiaq): Assume fdStream IO for closures
	default:
		ch.errorf(fn.Command, msg)
	}

	for _, rd := range fn.Redirs {
		if rd.Fd() <= 1 {
			annotation.streamTypes[rd.Fd()] = unusedStream
		}
	}

	if annotation.commandType == commandBuiltinSpecial {
		annotation.builtinSpecial.check(ch, fn)
	} else {
		ch.checkTermList(fn.Args)
	}
}

func (ch *Checker) checkTerms(tns []*parse.TermNode) {
	for _, tn := range tns {
		ch.checkTerm(tn)
	}
}

// checkTermList checks a TermListNode by checking all terms it contains.
func (ch *Checker) checkTermList(ln *parse.TermListNode) {
	ch.checkTerms(ln.Nodes)
}

// checkTerm checks a TermNode by checking all factors it contains.
func (ch *Checker) checkTerm(tn *parse.TermNode) {
	for _, fn := range tn.Nodes {
		ch.checkFactor(fn)
	}
}

// checkFactor checks a FactorNode by analyzing different factor types case by
// case. A StringFactor is not checked at all. A VariableFactor is resolved
// statically. The other composite factor types are checked recursively.
func (ch *Checker) checkFactor(fn *parse.FactorNode) {
	switch fn.Typ {
	case parse.StringFactor:
	case parse.VariableFactor:
		ch.resolveVar(fn.Node.(*parse.StringNode).Text, fn)
	case parse.TableFactor:
		table := fn.Node.(*parse.TableNode)
		for _, tn := range table.List {
			ch.checkTerm(tn)
		}
		for _, tp := range table.Dict {
			ch.checkTerm(tp.Key)
			ch.checkTerm(tp.Value)
		}
	case parse.ClosureFactor:
		ch.checkClosure(fn.Node.(*parse.ClosureNode))
	case parse.ListFactor:
		ch.checkTermList(fn.Node.(*parse.TermListNode))
	case parse.OutputCaptureFactor, parse.StatusCaptureFactor:
		ch.checkPipeline(fn.Node.(*parse.PipelineNode))
	}
}
