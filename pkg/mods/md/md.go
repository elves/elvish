// Package md exposes functionality from src.elv.sh/pkg/md.
package md

import (
	"src.elv.sh/pkg/elvdoc"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/md"
	"src.elv.sh/pkg/sys"
)

// Ns is the namespace for the md: module.
var Ns = eval.BuildNsNamed("md").
	AddGoFns(map[string]any{
		"show": show,
	}).Ns()

type showOpts struct {
	Width int
}

func (*showOpts) SetDefaultOptions() {}

func show(fm *eval.Frame, opts showOpts, markdown string) error {
	width := opts.Width
	if width <= 0 {
		_, width = sys.WinSize(fm.Port(1).File)
		if width <= 0 {
			width = 80
		}
	}
	codec := &md.TTYCodec{
		Width:              width,
		HighlightCodeBlock: elvdoc.HighlightCodeBlock,
	}
	_, err := fm.ByteOutput().WriteString(md.RenderString(markdown, codec))
	return err
}
