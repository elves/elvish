package main

import (
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/ui/styledown"
)

func Styledown(c etk.Context) (etk.View, etk.React) {
	codeView, codeReact := c.Subcomp("code",
		etk.WithInit(comps.TextArea, "prompt", ui.T("Styledown:\n")))
	content := etk.BindState(c, "code/buffer", comps.TextBuffer{}).Get().Content
	rendered, err := styledown.Render(content)
	if err == nil {
		rendered = ui.Concat(ui.T("Rendered:\n"), rendered)
	} else {
		rendered = ui.T("Error: "+err.Error(), ui.FgRed)
	}
	return etk.Box(`
		[code=]
		rendered=`, codeView, etk.Text(rendered)), codeReact
}
