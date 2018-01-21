package eval

import "github.com/elves/elvish/parse"

func (cp *compiler) chunkOp(n *parse.Chunk) Op {
	cp.compiling(n)
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
	cp.compiling(n)
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
	cp.compiling(n)
	return Op{cp.form(n), n.Begin(), n.End()}
}

func (cp *compiler) formOps(ns []*parse.Form) []Op {
	ops := make([]Op, len(ns))
	for i, n := range ns {
		ops[i] = cp.formOp(n)
	}
	return ops
}

func (cp *compiler) assignmentOp(n *parse.Assignment) Op {
	cp.compiling(n)
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
	cp.compiling(n)
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
	cp.compiling(n)
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
	cp.compiling(n)
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
	cp.compiling(n)
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
	cp.compiling(n)
	return ValuesOp{cp.primary(n), n.Begin(), n.End()}
}

func (cp *compiler) primaryOps(ns []*parse.Primary) []ValuesOp {
	ops := make([]ValuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.primaryOp(n)
	}
	return ops
}

func (cp *compiler) listOp(n *parse.Primary) ValuesOp {
	cp.compiling(n)
	return ValuesOp{cp.list(n), n.Begin(), n.End()}
}

func (cp *compiler) listOps(ns []*parse.Primary) []ValuesOp {
	ops := make([]ValuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.listOp(n)
	}
	return ops
}

func (cp *compiler) lambdaOp(n *parse.Primary) ValuesOp {
	cp.compiling(n)
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
	cp.compiling(n)
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
	cp.compiling(n)
	return ValuesOp{cp.braced(n), n.Begin(), n.End()}
}

func (cp *compiler) bracedOps(ns []*parse.Primary) []ValuesOp {
	ops := make([]ValuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.bracedOp(n)
	}
	return ops
}
