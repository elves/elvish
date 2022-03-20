package vals

import "reflect"

var (
	dummy              any
	nilValue           = reflect.ValueOf(&dummy).Elem()
	emptyInterfaceType = reflect.TypeOf(&dummy).Elem()
)

// ValueOf is like reflect.ValueOf, except that when given an argument of nil,
// it does not return a zero Value, but the Value for the zero value of the
// empty interface.
func ValueOf(i any) reflect.Value {
	if i == nil {
		return nilValue
	}
	return reflect.ValueOf(i)
}

// TypeOf is like reflect.TypeOf, except that when given an argument of nil, it
// does not return nil, but the Type for the empty interface.
func TypeOf(i any) reflect.Type {
	if i == nil {
		return emptyInterfaceType
	}
	return reflect.TypeOf(i)
}
