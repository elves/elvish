//go:build !unix

package unix

import (
	"src.elv.sh/pkg/eval"
)

// ExposeUnixNs indicate whether this module should be exposed as a usable
// elvish namespace.
const ExposeUnixNs = false

// Ns is an Elvish namespace that contains variables and functions that deal
// with features unique to Unix-like operating systems.
var Ns = &eval.Ns{}
