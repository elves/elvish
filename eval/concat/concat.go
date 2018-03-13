package concat

import (
	"fmt"
	"reflect"

	"github.com/elves/elvish/eval/vals"
)

type Concater = func(lhs, rhs interface{}) (interface{}, error)

var registry = make(map[reflect.Type]map[reflect.Type]Concater)

func Register(lhs, rhs reflect.Type, concater Concater) {
	submap, ok := registry[lhs]
	if !ok {
		submap = make(map[reflect.Type]Concater)
		registry[lhs] = submap
	}
	submap[rhs] = concater
}

func Concat(lhs, rhs interface{}) (interface{}, error) {
	if submap, ok := registry[reflect.TypeOf(lhs)]; ok {
		if concater, ok := submap[reflect.TypeOf(rhs)]; ok {
			return concater(lhs, rhs)
		}
	}

	return nil, fmt.Errorf("unsupported concat: %s and %s",
		vals.Kind(lhs), vals.Kind(rhs))
}

func init() {
	stringType := reflect.TypeOf((*string)(nil)).Elem()

	Register(stringType, stringType, func(lhs, rhs interface{}) (interface{}, error) {
		return lhs.(string) + rhs.(string), nil
	})
}
