package edit

import (
	"fmt"

	"github.com/elves/elvish/edit/history"
	"github.com/elves/elvish/edit/ui"
)

// Command history subsystem.

// Interface.

var _ = registerBuiltins("history", map[string]func(*Editor){
	"start":              historyStart,
	"up":                 historyUp,
	"down":               historyDown,
	"down-or-quit":       historyDownOrQuit,
	"switch-to-histlist": historySwitchToHistlist,
	"default":            historyDefault,
})

func init() {
	registerBindings(modeHistory, "history", map[ui.Key]string{
		{ui.Up, 0}:     "up",
		{ui.Down, 0}:   "down-or-quit",
		{'[', ui.Ctrl}: "insert:start",
		{'R', ui.Ctrl}: "switch-to-histlist",
		ui.Default:     "default",
	})
}

type hist struct {
	*history.Walker
}

func (hist) Mode() ModeType {
	return modeHistory
}

func (h *hist) ModeLine() renderer {
	return modeLineRenderer{fmt.Sprintf(" HISTORY #%d ", h.CurrentSeq()), ""}
}

func historyStart(ed *Editor) {
	prefix := ed.line[:ed.dot]
	walker := history.NewWalker(ed.store, -1, nil, nil, prefix)
	ed.hist = hist{walker}
	if ed.prevHistory() {
		ed.mode = &ed.hist
	} else {
		ed.addTip("no matching history item")
	}
}

func historyUp(ed *Editor) {
	ed.prevHistory()
}

func historyDown(ed *Editor) {
	ed.nextHistory()
}

func historyDownOrQuit(ed *Editor) {
	if !ed.nextHistory() {
		ed.mode = &ed.insert
	}
}

func historySwitchToHistlist(ed *Editor) {
	histlistStart(ed)
	if ed.mode == ed.histlist {
		ed.line = ""
		ed.dot = 0
		ed.histlist.changeFilter(ed.hist.Prefix())
	}
}

func historyDefault(ed *Editor) {
	ed.acceptHistory()
	ed.mode = &ed.insert
	ed.nextAction = action{typ: reprocessKey}
}

// Implementation.

func (ed *Editor) appendHistory(line string) {
	if ed.store != nil {
		ed.historyMutex.Lock()
		ed.store.Waits.Add(1)
		go func() {
			// TODO(xiaq): Report possible error
			_, err := ed.store.AddCmd(line)
			ed.store.Waits.Done()
			ed.historyMutex.Unlock()
			if err != nil {
				logger.Println("failed to add cmd to store:", err)
			} else {
				logger.Println("added cmd to store:", line)
			}
		}()
	}
}

func (ed *Editor) prevHistory() bool {
	_, _, err := ed.hist.Prev()
	return err == nil
}

func (ed *Editor) nextHistory() bool {
	_, _, err := ed.hist.Next()
	return err == nil
}

// acceptHistory accepts the currently selected history.
func (ed *Editor) acceptHistory() {
	ed.line = ed.hist.CurrentCmd()
	ed.dot = len(ed.line)
}
