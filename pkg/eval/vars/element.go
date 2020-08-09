package vars

import (
	"github.com/elves/elvish/pkg/eval/vals"
)

type elem struct {
	variable Var
	assocers []interface{}
	indices  []interface{}
	setValue interface{}
}

func (ev *elem) Set(v0 interface{}) error {
	var err error
	v := v0
	// Evaluate the actual new value from inside out. See comments in
	// MakeElement for how element assignment works.
	for i := len(ev.assocers) - 1; i >= 0; i-- {
		v, err = vals.Assoc(ev.assocers[i], ev.indices[i], v)
		if err != nil {
			return err
		}
	}
	err = ev.variable.Set(v)
	// TODO(xiaq): Remember the set value for use in Get.
	ev.setValue = v0
	return err
}

func (ev *elem) Get() interface{} {
	// TODO(xiaq): This is only called from fixNilVariables. We don't want to
	// waste time accessing the variable, so we simply return the value that was
	// set.
	return ev.setValue
}

// MakeElement returns a variable, that when set, simulates the mutation of an
// element.
func MakeElement(v Var, indicies []interface{}) (Var, error) {
	// Assignment of indexed variables actually assignes the variable, with
	// the right hand being a nested series of Assocs. As the simplest
	// example, `a[0] = x` is equivalent to `a = (assoc $a 0 x)`. A more
	// complex example is that `a[0][1][2] = x` is equivalent to
	//	`a = (assoc $a 0 (assoc $a[0] 1 (assoc $a[0][1] 2 x)))`.
	// Note that in each assoc form, the first two arguments can be
	// determined now, while the last argument is only known when the
	// right-hand-side is known. So here we evaluate the first two arguments
	// of each assoc form and put them in two slices, assocers and indicies.
	// In the previous example, the two slices will contain:
	//
	// assocers: $a $a[0] $a[0][1]
	// indicies:  0     1        2
	//
	// When the right-hand side of the assignment becomes available, the new
	// value for $a is evaluated by doing Assoc from inside out.
	assocers := make([]interface{}, len(indicies))
	varValue := v.Get()
	assocers[0] = varValue
	for i, index := range indicies[:len(indicies)-1] {
		lastAssocer := assocers[i]
		v, err := vals.Index(lastAssocer, index)
		if err != nil {
			return nil, err
		}
		assocers[i+1] = v
	}
	return &elem{v, assocers, indicies, nil}, nil
}

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

type elemErr struct {
	level int
	msg   string
}

func (err elemErr) Error() string {
	return err.msg
}

// HeadOfElement gets the underlying head variable of an element variable, or
// nil if the argument is not an element variable.
func HeadOfElement(v Var) Var {
	if ev, ok := v.(*elem); ok {
		return ev.variable
	}
	return nil
}

// ElementErrorLevel returns the level of an error returned by MakeElement or
// DelElement. Level 0 represents that the error is about the variable itself.
// If the argument was not returned from MakeVariable, -1 is returned.
func ElementErrorLevel(err error) int {
	if err, ok := err.(elemErr); ok {
		return err.level
	}
	return -1
}
