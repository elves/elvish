// Command widget allows manually testing a single widget.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/ui"
)

var (
	maxHeight  = flag.Int("max-height", 10, "maximum height")
	horizontal = flag.Bool("horizontal", false, "use horizontal listbox layout")
)

func makeWidget() tk.Widget {
	items := tk.TestItems{Prefix: "list item "}
	w := tk.NewComboBox(tk.ComboBoxSpec{
		CodeArea: tk.CodeAreaSpec{
			Prompt: func() ui.Text {
				return ui.Concat(ui.T(" NUMBER ", ui.Bold, ui.BgMagenta), ui.T(" "))
			},
		},
		ListBox: tk.ListBoxSpec{
			State:       tk.ListBoxState{Items: &items},
			Placeholder: ui.T("(no items)"),
			Horizontal:  *horizontal,
		},
		OnFilter: func(w tk.ComboBox, filter string) {
			if n, err := strconv.Atoi(filter); err == nil {
				items.NItems = n
			}
		},
	})
	return w
}

func main() {
	flag.Parse()
	widget := makeWidget()

	tty := cli.NewTTY(os.Stdin, os.Stderr)
	restore, err := tty.Setup()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer restore()
	defer tty.CloseReader()
	for {
		h, w := tty.Size()
		if h > *maxHeight {
			h = *maxHeight
		}
		tty.UpdateBuffer(nil, widget.Render(w, h), false)
		event, err := tty.ReadEvent()
		if err != nil {
			errBuf := term.NewBufferBuilder(w).Write(err.Error(), ui.FgRed).Buffer()
			tty.UpdateBuffer(nil, errBuf, true)
			break
		}
		handled := widget.Handle(event)
		if !handled && event == term.K('D', ui.Ctrl) {
			tty.UpdateBuffer(nil, term.NewBufferBuilder(w).Buffer(), true)
			break
		}
	}
}
