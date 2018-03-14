package concat

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/elves/elvish/eval/vals"
)

type (
	conc   func(lhs, rhs interface{}) (interface{}, error)
	matrix map[reflect.Type]map[reflect.Type]conc
)

var reg = struct {
	sync.RWMutex
	m matrix
}{m: make(matrix)}

func Register(lhs, rhs reflect.Type, concater conc) {
	reg.Lock()
	defer reg.Unlock()

	submap, ok := reg.m[lhs]
	if !ok {
		submap = make(map[reflect.Type]conc)
		reg.m[lhs] = submap
	}
	submap[rhs] = concater
}

func Concat(lhs, rhs interface{}) (interface{}, error) {
	reg.RLock()
	defer reg.RUnlock()

	if concater, ok := reg.m[reflect.TypeOf(lhs)][reflect.TypeOf(rhs)]; ok {
		return concater(lhs, rhs)
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
