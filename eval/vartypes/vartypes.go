// Package vartypes contains basic types for manipulating Elvish variables.
package vartypes

// Variable represents an Elvish variable.
type Variable interface {
	Set(v interface{}) error
	Get() interface{}
}
