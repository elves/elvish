package eval

import "github.com/elves/elvish/parse"

func (cp *compiler) chunkOp(n *parse.Chunk) effectOp {
	cp.compiling(n)
	return effectOp{cp.chunk(n), n.Range().From, n.Range().To}
}

func (cp *compiler) chunkOps(ns []*parse.Chunk) []effectOp {
	ops := make([]effectOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.chunkOp(n)
	}
	return ops
}

func (cp *compiler) pipelineOp(n *parse.Pipeline) effectOp {
	cp.compiling(n)
	return effectOp{cp.pipeline(n), n.Range().From, n.Range().To}
}

func (cp *compiler) pipelineOps(ns []*parse.Pipeline) []effectOp {
	ops := make([]effectOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.pipelineOp(n)
	}
	return ops
}

func (cp *compiler) formOp(n *parse.Form) effectOp {
	cp.compiling(n)
	return effectOp{cp.form(n), n.Range().From, n.Range().To}
}

func (cp *compiler) formOps(ns []*parse.Form) []effectOp {
	ops := make([]effectOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.formOp(n)
	}
	return ops
}

func (cp *compiler) assignmentOp(n *parse.Assignment, temporary bool) effectOp {
	cp.compiling(n)
	return effectOp{cp.assignment(n, temporary), n.Range().From, n.Range().To}
}

func (cp *compiler) assignmentOps(ns []*parse.Assignment, temporary bool) []effectOp {
	ops := make([]effectOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.assignmentOp(n, temporary)
	}
	return ops
}

func (cp *compiler) redirOp(n *parse.Redir) effectOp {
	cp.compiling(n)
	return effectOp{cp.redir(n), n.Range().From, n.Range().To}
}

func (cp *compiler) redirOps(ns []*parse.Redir) []effectOp {
	ops := make([]effectOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.redirOp(n)
	}
	return ops
}

func (cp *compiler) compoundOp(n *parse.Compound) valuesOp {
	cp.compiling(n)
	return valuesOp{cp.compound(n), n.Range().From, n.Range().To}
}

func (cp *compiler) compoundOps(ns []*parse.Compound) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.compoundOp(n)
	}
	return ops
}

func (cp *compiler) arrayOp(n *parse.Array) valuesOp {
	cp.compiling(n)
	return valuesOp{cp.array(n), n.Range().From, n.Range().To}
}

func (cp *compiler) arrayOps(ns []*parse.Array) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.arrayOp(n)
	}
	return ops
}

func (cp *compiler) indexingOp(n *parse.Indexing) valuesOp {
	cp.compiling(n)
	return valuesOp{cp.indexing(n), n.Range().From, n.Range().To}
}

func (cp *compiler) indexingOps(ns []*parse.Indexing) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.indexingOp(n)
	}
	return ops
}

func (cp *compiler) primaryOp(n *parse.Primary) valuesOp {
	cp.compiling(n)
	return valuesOp{cp.primary(n), n.Range().From, n.Range().To}
}

func (cp *compiler) primaryOps(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.primaryOp(n)
	}
	return ops
}

func (cp *compiler) listOp(n *parse.Primary) valuesOp {
	cp.compiling(n)
	return valuesOp{cp.list(n), n.Range().From, n.Range().To}
}

func (cp *compiler) listOps(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.listOp(n)
	}
	return ops
}

func (cp *compiler) lambdaOp(n *parse.Primary) valuesOp {
	cp.compiling(n)
	return valuesOp{cp.lambda(n), n.Range().From, n.Range().To}
}

func (cp *compiler) lambdaOps(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.lambdaOp(n)
	}
	return ops
}

func (cp *compiler) map_Op(n *parse.Primary) valuesOp {
	cp.compiling(n)
	return valuesOp{cp.map_(n), n.Range().From, n.Range().To}
}

func (cp *compiler) map_Ops(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.map_Op(n)
	}
	return ops
}

func (cp *compiler) bracedOp(n *parse.Primary) valuesOp {
	cp.compiling(n)
	return valuesOp{cp.braced(n), n.Range().From, n.Range().To}
}

func (cp *compiler) bracedOps(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.bracedOp(n)
	}
	return ops
}
