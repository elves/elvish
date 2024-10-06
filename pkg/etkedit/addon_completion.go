package edit

import (
	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/edit/complete"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/ui"
)

func startCompletion(ed *Editor, c etk.Context) {
	bufferVar := bufferVar(c)
	b := bufferVar.Get()
	result, err := complete.Complete(
		complete.CodeBuffer{Content: b.Content, Dot: b.Dot},
		c.Frame().Evaler,
		// TODO
		complete.Config{})
	if err != nil {
		c.AddMsg(errorMsgText(err))
		return
	}
	items := completionItems(result.Items)
	pendingVar := pendingVar(c)
	updatePending := func(item modes.CompletionItem) {
		pendingVar.Set(comps.PendingText{
			From:    result.Replace.From,
			To:      result.Replace.To,
			Content: item.ToInsert,
		})
	}
	if len(items) > 0 {
		updatePending(items[0])
	}
	pushAddonWithDismiss(c, withAfterReact(
		etk.WithInit(comps.ComboBox,
			"query/prompt", addonPromptText(" COMPLETE "),
			"list/multi-column", true,
			"gen-list", func(f string) (comps.ListItems, int) {
				// TODO: Implement actual filtering
				if len(items) > 0 {
					updatePending(items[0])
				}
				return items, 0
			},
			"binding", etkBindingFromBindingMap(ed, &ed.completionBinding),
		),
		func(c etk.Context, r etk.Reaction) etk.Reaction {
			if r == etk.Finish {
				// TODO: This should just be a call to comps.ApplyPending with a
				// child Context
				bufferVar.Swap(func(buf comps.TextBuffer) comps.TextBuffer {
					buf, _, _ = comps.PatchPending(buf, pendingVar.Get())
					return buf
				})
				return etk.Finish
			}
			items := etk.BindState(c, "list/items", comps.ListItems(nil)).Get()
			selected := etk.BindState(c, "list/selected", 0).Get()
			if items != nil && 0 <= selected && selected < items.Len() {
				item, ok := items.Get(selected).(modes.CompletionItem)
				if ok {
					pendingVar.Set(comps.PendingText{
						From:    result.Replace.From,
						To:      result.Replace.To,
						Content: item.ToInsert,
					})
				}
			}
			return r
		}),
		true,
		func() { pendingVar.Set(comps.PendingText{}) },
	)
}

type completionItems []modes.CompletionItem

func (ci completionItems) Len() int           { return len(ci) }
func (ci completionItems) Get(i int) any      { return ci[i] }
func (ci completionItems) Show(i int) ui.Text { return ci[i].ToShow }
