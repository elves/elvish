package edcore

import (
	"github.com/elves/elvish/eval"
)

// This file contains utilities that facilitates modularization of the editor.

var editorInitFuncs []func(*editor, eval.Ns)

func atEditorInit(f func(*editor, eval.Ns)) {
	editorInitFuncs = append(editorInitFuncs, f)
}
