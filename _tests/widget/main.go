// Command widget allows manually testing a single widget.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/combobox"
	"github.com/elves/elvish/cli/listbox"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

// Change this value to test another widget.
var widget clitypes.Widget = makeCombobox()

func makeCombobox() clitypes.Widget {
	w := &combobox.Widget{
		ListBox: listbox.Widget{State: listbox.State{Itemer: itemer{}}},
	}
	w.OnFilter = func(filter string) {
		n, err := strconv.Atoi(filter)
		if filter == "" {
			n, err = 100, nil
		}
		if err == nil {
			w.ListBox.MutateListboxState(func(s *listbox.State) {
				s.NItems = n
			})
		}
	}
	return w
}

var maxHeight = 10

type itemer struct{}

func (it itemer) Item(i int) styled.Text { return styled.Plain(strconv.Itoa(i)) }

func main() {
	tty := clicore.NewTTY(os.Stdin, os.Stderr)
	restore, err := tty.Setup()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer restore()
	events := tty.StartInput()
	defer tty.StopInput()
	for {
		h, w := tty.Size()
		if h > maxHeight {
			h = maxHeight
		}
		tty.UpdateBuffer(nil, widget.Render(w, h), false)
		event := <-events
		handled := widget.Handle(event)
		if !handled && event == term.K('D', ui.Ctrl) {
			tty.UpdateBuffer(nil, ui.NewBufferBuilder(w).Buffer(), true)
			break
		}
	}
}
