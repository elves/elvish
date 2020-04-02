// +build !windows,!plan9,!js

// Package unix exposes variables and functions that deal with features
// unique to UNIX-like operating systems. On non-UNIX operating systems it
// will be an empty namespace.
package unix

import (
	"github.com/elves/elvish/pkg/eval"
)

// Indicate that this module should be exposed as a usable elvish namespace.
const ExposeUnixNs = true

var Ns = eval.Ns{
	"umask": UmaskVariable{},
}
