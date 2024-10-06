package main

import (
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/ui/styledown"
)

func Styledown(c etk.Context) (etk.View, etk.React) {
	codeView, codeReact := c.Subcomp("code",
		etk.WithInit(etk.TextArea, "prompt", ui.T("Styledown:\n")))
	content := etk.BindState(c, "code/buffer", etk.TextBuffer{}).Get().Content
	rendered, err := styledown.Render(content)
	if err == nil {
		rendered = ui.Concat(ui.T("Rendered:\n"), rendered)
	} else {
		rendered = ui.T("Error: "+err.Error(), ui.FgRed)
	}
	return etk.VBoxView(0, codeView, etk.TextView(0, rendered)), codeReact
}
