package vartypes

import (
	"errors"
	"strconv"

	"github.com/elves/elvish/eval/types"
)

var errMustBeNumber = errors.New("must be number")

type numberVar struct {
	ptr *float64
}

func NewNumber(ptr *float64) Variable {
	return numberVar{ptr}
}

func (nv numberVar) Get() types.Value {
	return types.String(strconv.FormatFloat(*nv.ptr, 'E', -1, 64))
}

func (nv numberVar) Set(v types.Value) error {
	if s, ok := v.(types.String); ok {
		if num, err := strconv.ParseFloat(string(s), 64); err == nil {
			*nv.ptr = num
			return nil
		}
	}
	return errMustBeNumber
}
