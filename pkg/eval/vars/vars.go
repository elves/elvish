// Package vars contains basic types for manipulating Elvish variables.
package vars

// Var represents an Elvish variable.
type Var interface {
	Set(v any) error
	Get() any
}

// UnsettableVar represents an Elvish variable that can be in an unset state.
type UnsettableVar interface {
	Var
	Unset() error
	IsSet() bool
}
