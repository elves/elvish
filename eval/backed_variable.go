package eval

import (
	"reflect"
)

type backedVariable struct {
	ptr interface{}
}

var _ Variable = backedVariable{}

var valueTypeReflect = reflect.TypeOf((*Value)(nil)).Elem()

// NewBackedVariable creates a variable backed by a pointer to something of a
// Value-compatible type.
//
// A Value-compatible type can be converted to and from a Value. The conversion
// to a Value is always safe, while the conversion from a Value might throw an
// Exception if an incompatible value is given. Currently the following types
// are supported:
//
// 1) string, treated as the same as the String type. That is, conversion of a
//    string s to Value is String(s), and conversion from Value succeeds as long
//    as the Value has type String.
// 2) int. Converts to a String; conversion from Value is defined by toInt, with
//    error turned into Exception.
// 3) float64. Converts to a String; conversion from Value is defined by
//    toFloat, with error turned into Exception.
// 4) Any type that implements Value. Conversion to Value is a safe type cast,
//    and conversion from Value is a type assertion, with failed assertion
//    converted to an Exception.
//
// The function panics if the argument is not a pointer to a Value-compatible
// value.
func NewBackedVariable(ptr interface{}) Variable {
	ptrReflect := reflect.ValueOf(ptr)
	if ptrReflect.Kind() != reflect.Ptr {
		panic("internal bug: NewBackedVariable only accepts pointer")
	}
	vReflect := reflect.Indirect(ptrReflect)
	k := vReflect.Kind()
	if !(k == reflect.Int || k == reflect.Float64 || k == reflect.String ||
		vReflect.Type().Implements(valueTypeReflect)) {
		panic("internal bug: NewBackedVariable only accepts pointer to Value-compatible type")
	}
	return backedVariable{ptr}
}

func (v backedVariable) Get() Value {
	return convertGoToValue(reflect.Indirect(reflect.ValueOf(v.ptr)).Interface())
}

func (v backedVariable) Set(newval Value) {
	scanValueToGo(newval, v.ptr)
}
