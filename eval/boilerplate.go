package eval

import "github.com/elves/elvish/parse"

func (cp *compiler) chunkOp(n *parse.Chunk) Op {
	return Op{cp.chunk(n), n.Begin(), n.End()}
}

func (cp *compiler) chunkOps(ns []*parse.Chunk) []Op {
	ops := make([]Op, len(ns))
	for i, n := range ns {
		ops[i] = cp.chunkOp(n)
	}
	return ops
}

func (cp *compiler) pipelineOp(n *parse.Pipeline) Op {
	return Op{cp.pipeline(n), n.Begin(), n.End()}
}

func (cp *compiler) pipelineOps(ns []*parse.Pipeline) []Op {
	ops := make([]Op, len(ns))
	for i, n := range ns {
		ops[i] = cp.pipelineOp(n)
	}
	return ops
}

func (cp *compiler) formOp(n *parse.Form) Op {
	return Op{cp.form(n), n.Begin(), n.End()}
}

func (cp *compiler) formOps(ns []*parse.Form) []Op {
	ops := make([]Op, len(ns))
	for i, n := range ns {
		ops[i] = cp.formOp(n)
	}
	return ops
}

func (cp *compiler) controlOp(n *parse.Control) Op {
	return Op{cp.control(n), n.Begin(), n.End()}
}

func (cp *compiler) controlOps(ns []*parse.Control) []Op {
	ops := make([]Op, len(ns))
	for i, n := range ns {
		ops[i] = cp.controlOp(n)
	}
	return ops
}

func (cp *compiler) assignmentOp(n *parse.Assignment) Op {
	return Op{cp.assignment(n), n.Begin(), n.End()}
}

func (cp *compiler) assignmentOps(ns []*parse.Assignment) []Op {
	ops := make([]Op, len(ns))
	for i, n := range ns {
		ops[i] = cp.assignmentOp(n)
	}
	return ops
}

func (cp *compiler) redirOp(n *parse.Redir) Op {
	return Op{cp.redir(n), n.Begin(), n.End()}
}

func (cp *compiler) redirOps(ns []*parse.Redir) []Op {
	ops := make([]Op, len(ns))
	for i, n := range ns {
		ops[i] = cp.redirOp(n)
	}
	return ops
}

func (cp *compiler) compoundOp(n *parse.Compound) ValuesOp {
	return ValuesOp{cp.compound(n), n.Begin(), n.End()}
}

func (cp *compiler) compoundOps(ns []*parse.Compound) []ValuesOp {
	ops := make([]ValuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.compoundOp(n)
	}
	return ops
}

func (cp *compiler) arrayOp(n *parse.Array) ValuesOp {
	return ValuesOp{cp.array(n), n.Begin(), n.End()}
}

func (cp *compiler) arrayOps(ns []*parse.Array) []ValuesOp {
	ops := make([]ValuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.arrayOp(n)
	}
	return ops
}

func (cp *compiler) indexingOp(n *parse.Indexing) ValuesOp {
	return ValuesOp{cp.indexing(n), n.Begin(), n.End()}
}

func (cp *compiler) indexingOps(ns []*parse.Indexing) []ValuesOp {
	ops := make([]ValuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.indexingOp(n)
	}
	return ops
}

func (cp *compiler) primaryOp(n *parse.Primary) ValuesOp {
	return ValuesOp{cp.primary(n), n.Begin(), n.End()}
}

func (cp *compiler) primaryOps(ns []*parse.Primary) []ValuesOp {
	ops := make([]ValuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.primaryOp(n)
	}
	return ops
}

func (cp *compiler) errorCaptureOp(n *parse.Chunk) ValuesOp {
	return ValuesOp{cp.errorCapture(n), n.Begin(), n.End()}
}

func (cp *compiler) errorCaptureOps(ns []*parse.Chunk) []ValuesOp {
	ops := make([]ValuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.errorCaptureOp(n)
	}
	return ops
}

func (cp *compiler) outputCaptureOp(n *parse.Primary) ValuesOp {
	return ValuesOp{cp.outputCapture(n), n.Begin(), n.End()}
}

func (cp *compiler) outputCaptureOps(ns []*parse.Primary) []ValuesOp {
	ops := make([]ValuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.outputCaptureOp(n)
	}
	return ops
}

func (cp *compiler) lambdaOp(n *parse.Primary) ValuesOp {
	return ValuesOp{cp.lambda(n), n.Begin(), n.End()}
}

func (cp *compiler) lambdaOps(ns []*parse.Primary) []ValuesOp {
	ops := make([]ValuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.lambdaOp(n)
	}
	return ops
}

func (cp *compiler) map_Op(n *parse.Primary) ValuesOp {
	return ValuesOp{cp.map_(n), n.Begin(), n.End()}
}

func (cp *compiler) map_Ops(ns []*parse.Primary) []ValuesOp {
	ops := make([]ValuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.map_Op(n)
	}
	return ops
}

func (cp *compiler) bracedOp(n *parse.Primary) ValuesOp {
	return ValuesOp{cp.braced(n), n.Begin(), n.End()}
}

func (cp *compiler) bracedOps(ns []*parse.Primary) []ValuesOp {
	ops := make([]ValuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.bracedOp(n)
	}
	return ops
}
