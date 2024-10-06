package edit

import "src.elv.sh/pkg/ui"

func addonPromptText(text string) ui.Text {
	return ui.Concat(ui.T(text, ui.Bold, ui.FgWhite, ui.BgMagenta), ui.T(" "))
}

func errorMsgText(e error) ui.Text {
	return ui.Concat(ui.T("error: ", ui.FgRed), ui.T(e.Error()))
}
