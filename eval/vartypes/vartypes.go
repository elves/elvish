// Package vartypes contains basic types for manipulating Elvish variables.
package vartypes

import (
	"github.com/elves/elvish/eval/types"
)

// Variable represents an Elvish variable.
type Variable interface {
	Set(v types.Value) error
	Get() types.Value
}
