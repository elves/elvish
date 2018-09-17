// Package edit implements the line editor for Elvish.
//
// The line editor is organized into a core package, which is a general,
// Elvish-agnostic line editor, and multiple "addon" packages that implement
// things like asynchronous prompt, code highlighting, various modes along with
// their Elvish APIs. This package assembles all those packages.
package newedit
