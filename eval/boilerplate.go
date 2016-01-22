package eval

import "github.com/elves/elvish/parse-ng"

func (cp *compiler) chunks(ns []*parse.Chunk) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.chunk(n)
	}
	return ops
}

func (cp *compiler) pipelines(ns []*parse.Pipeline) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.pipeline(n)
	}
	return ops
}

func (cp *compiler) forms(ns []*parse.Form) []stateUpdatesOp {
	ops := make([]stateUpdatesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.form(n)
	}
	return ops
}

func (cp *compiler) redirs(ns []*parse.Redir) []op {
	ops := make([]op, len(ns))
	for i, n := range ns {
		ops[i] = cp.redir(n)
	}
	return ops
}

func (cp *compiler) compounds(ns []*parse.Compound) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.compound(n)
	}
	return ops
}

func (cp *compiler) arrays(ns []*parse.Array) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.array(n)
	}
	return ops
}

func (cp *compiler) indexeds(ns []*parse.Indexed) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.indexed(n)
	}
	return ops
}

func (cp *compiler) primarys(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.primary(n)
	}
	return ops
}
