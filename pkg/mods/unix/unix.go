//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

// Package unix exports an Elvish namespace that contains variables and
// functions that deal with features unique to UNIX-like operating systems. On
// non-UNIX operating systems it exports an empty namespace.
package unix

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/logutil"
)

// ExposeUnixNs indicate whether this module should be exposed as a usable
// elvish namespace.
const ExposeUnixNs = true

// Ns is an Elvish namespace that contains variables and functions that deal
// with features unique to UNIX-like operating systems. On
var Ns = eval.BuildNs().
	AddVars(map[string]vars.Var{
		"umask":   UmaskVariable{},
		"rlimits": rlimitsVar{},
	}).Ns()

var logger = logutil.GetLogger("[mods/unix] ")
