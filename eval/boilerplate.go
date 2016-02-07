package eval

import "github.com/elves/elvish/parse"

func (cp *compiler) chunks(ns []*parse.Chunk) []op {
	ops := make([]op, len(ns))
	for i, n := range ns {
		ops[i] = cp.chunk(n)
	}
	return ops
}
func chunkBegins(ns []*parse.Chunk) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) pipelines(ns []*parse.Pipeline) []op {
	ops := make([]op, len(ns))
	for i, n := range ns {
		ops[i] = cp.pipeline(n)
	}
	return ops
}
func pipelineBegins(ns []*parse.Pipeline) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) forms(ns []*parse.Form) []op {
	ops := make([]op, len(ns))
	for i, n := range ns {
		ops[i] = cp.form(n)
	}
	return ops
}
func formBegins(ns []*parse.Form) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) assignments(ns []*parse.Assignment) []op {
	ops := make([]op, len(ns))
	for i, n := range ns {
		ops[i] = cp.assignment(n)
	}
	return ops
}
func assignmentBegins(ns []*parse.Assignment) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) indexingVars(ns []*parse.Indexing, msg string) []variableOp {
	ops := make([]variableOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.indexingVar(n, msg)
	}
	return ops
}
func indexingVarBegins(ns []*parse.Indexing) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) redirs(ns []*parse.Redir) []op {
	ops := make([]op, len(ns))
	for i, n := range ns {
		ops[i] = cp.redir(n)
	}
	return ops
}
func redirBegins(ns []*parse.Redir) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) compounds(ns []*parse.Compound) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.compound(n)
	}
	return ops
}
func compoundBegins(ns []*parse.Compound) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) arrays(ns []*parse.Array) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.array(n)
	}
	return ops
}
func arrayBegins(ns []*parse.Array) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) indexings(ns []*parse.Indexing) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.indexing(n)
	}
	return ops
}
func indexingBegins(ns []*parse.Indexing) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) primarys(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.primary(n)
	}
	return ops
}
func primaryBegins(ns []*parse.Primary) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) outputCaptures(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.outputCapture(n)
	}
	return ops
}
func outputCaptureBegins(ns []*parse.Primary) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) lambdas(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.lambda(n)
	}
	return ops
}
func lambdaBegins(ns []*parse.Primary) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) map_s(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.map_(n)
	}
	return ops
}
func map_Begins(ns []*parse.Primary) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}

func (cp *compiler) braceds(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.braced(n)
	}
	return ops
}
func bracedBegins(ns []*parse.Primary) []int {
	begins := make([]int, len(ns))
	for i, n := range ns {
		begins[i] = n.Begin()
	}
	return begins
}
