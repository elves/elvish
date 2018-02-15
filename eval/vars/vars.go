// Package vars contains basic types for manipulating Elvish variables.
package vars

// Type represents an Elvish variable.
type Type interface {
	Set(v interface{}) error
	Get() interface{}
}
