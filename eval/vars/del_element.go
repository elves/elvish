package vars

import "github.com/elves/elvish/eval/vals"

// DelElement deletes an element. It uses a similar process to MakeElement,
// except that the last level of container needs to be Dissoc-able instead of
// Assoc-able.
func DelElement(variable Var, indicies []interface{}) error {
	var err error
	// In "del a[0][1][2]",
	//
	// indicies:  0  1     2
	// assocers: $a $a[0]
	// dissocer:          $a[0][1]
	assocers := make([]interface{}, len(indicies)-1)
	container := variable.Get()
	for i, index := range indicies[:len(indicies)-1] {
		assocers[i] = container

		var err error
		container, err = vals.Index(container, index)
		if err != nil {
			return err
		}
	}

	v := vals.Dissoc(container, indicies[len(indicies)-1])
	if v == nil {
		return elemErr{len(indicies), "value does not support element removal"}
	}

	for i := len(assocers) - 1; i >= 0; i-- {
		v, err = vals.Assoc(assocers[i], indicies[i], v)
		if err != nil {
			return err
		}
	}
	return variable.Set(v)
}
