// Package vars contains basic types for manipulating Elvish variables.
package vars

// Var represents an Elvish variable.
type Var interface {
	Set(v interface{}) error
	Get() interface{}
}
