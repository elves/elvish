package vartypes

import (
	"errors"
	"strconv"
)

var errMustBeNumber = errors.New("must be number")

type number struct {
	ptr *float64
}

func NewNumber(ptr *float64) Variable {
	return number{ptr}
}

func (nv number) Get() interface{} {
	return string(strconv.FormatFloat(*nv.ptr, 'E', -1, 64))
}

func (nv number) Set(v interface{}) error {
	if s, ok := v.(string); ok {
		if num, err := strconv.ParseFloat(string(s), 64); err == nil {
			*nv.ptr = num
			return nil
		}
	}
	return errMustBeNumber
}
