// Package pkg is the root of packages that implement Elvish.
package pkg

import "embed"

// ElvFiles contains the Elvish sources found inside pkg - the .d.elv files for
// modules implemented in Go and the actual sources of bundled modules. It is
// defined to support reading documentation of builtin modules from Elvish
// itself.
//
// Some of these files may be embedded as a string variable elsewhere. This is
// fine since the compiler is smart enough to only include one copy of the same
// file in the binary (as least as of Go 1.21).
//
// This is only used by [src.elv.sh/pkg/mods/doc], but has to live in this
// package because go:embed only supports embedding files found in the current
// directory and its descendents, and this directory is the lowest common
// ancestor of these files.
//
//go:embed eval/*.elv edit/*.elv mods/*/*.elv
var ElvFiles embed.FS
