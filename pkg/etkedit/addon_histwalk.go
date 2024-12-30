package edit

import (
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
)

func startHistwalk(ed *Editor, c etk.Context) {
	bufferVar := bufferVar(c)

	buffer := bufferVar.Get()
	prefix := buffer.Content[:buffer.Dot]
	cursor := ed.histStore.Cursor(prefix)
	cursor.Prev()
	cmd, err := cursor.Get()
	if err != nil {
		// TODO: Report error
		return
	}

	pendingVar := pendingVar(c)
	pendingVar.Set(comps.PendingText{
		From: len(prefix), To: len(buffer.Content),
		Content: cmd.Text[len(prefix):],
	})

	pushAddonWithDismiss(c, withAfterReact(
		etk.WithInit(comps.TextArea, "buffer", comps.TextBuffer{Content: "walk"}),
		func(c etk.Context, r etk.Reaction) etk.Reaction {
			// TODO: Handle up/down
			return r
		}),
		1,
		func() { pendingVar.Set(comps.PendingText{}) },
	)
}
