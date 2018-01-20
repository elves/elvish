package vartypes

import "github.com/elves/elvish/eval/types"

// DelElement deletes an element. It uses a similar process to MakeElement,
// except that the last level of container needs to be a Dissocer instead of an
// Assocer.
func DelElement(variable Variable, indicies []types.Value) error {
	// In "del a[0][1][2]",
	//
	// indicies:  0  1     2
	// assocers: $a $a[0]
	// dissocer:          $a[0][1]
	assocers := make([]types.Assocer, len(indicies)-1)
	container := variable.Get()
	for i, index := range indicies[:len(indicies)-1] {
		indexer, ok := container.(types.Indexer)
		if !ok {
			return elemErr{i, "value does not support indexing"}
		}
		assocer, ok := container.(types.Assocer)
		if !ok {
			return elemErr{i, "value does not support indexing for setting"}
		}
		assocers[i] = assocer

		var err error
		container, err = indexer.Index(index)
		if err != nil {
			return err
		}
	}
	dissocer, ok := container.(types.Dissocer)
	if !ok {
		return elemErr{len(indicies), "value does not support element removal"}
	}

	v := dissocer.Dissoc(indicies[len(indicies)-1])
	for i := len(assocers) - 1; i >= 0; i-- {
		v = assocers[i].Assoc(indicies[i], v)
	}
	return variable.Set(v)
}
