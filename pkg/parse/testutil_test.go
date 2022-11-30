package parse

import (
	"reflect"

	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

var nodeType = reflect.TypeOf((*Node)(nil)).Elem()
