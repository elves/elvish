//File module which is an equivalent of POSIX like commands.
//Checkout issule number 1263 for more information.
package file

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
)

var Ns = eval.NsBuilder{}.AddGoFns("file:", fns).Ns()

var fns = map[string]interface{}{}
