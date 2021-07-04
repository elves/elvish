//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

// Package unix exports an Elvish namespace that contains variables and
// functions that deal with features unique to UNIX-like operating systems. On
// non-UNIX operating systems it exports an empty namespace.
package unix

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vars"
)

// ExposeUnixNs indicate whether this module should be exposed as a usable
// elvish namespace.
const ExposeUnixNs = true

// Ns is an Elvish namespace that contains variables and functions that deal
// with features unique to UNIX-like operating systems. On
var Ns = eval.BuildNsNamed("unix").
	AddVars(map[string]vars.Var{
		"umask": UmaskVariable{},
	}).
	AddGoFns(map[string]interface{}{
		"ulimit": ulimit,
	}).Ns()
