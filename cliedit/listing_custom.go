package cliedit

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/addons/listing"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/ui"
)

type customListingOpts struct {
	Binding    BindingMap
	Caption    string
	KeepBottom bool
	Accept     eval.Callable
	AutoAccept bool
}

func (*customListingOpts) SetDefaultOptions() {}

//elvdoc:fn listing:start-custom
//
// Starts a custom listing addon.

func listingStartCustom(app cli.App, fm *eval.Frame, opts customListingOpts, items interface{}) {
	var binding el.Handler
	if opts.Binding.Map != nil {
		binding = newMapBinding(app, fm.Evaler, vars.FromPtr(&opts.Binding))
	}
	var getItems func(string) []listing.Item
	if fn, isFn := items.(eval.Callable); isFn {
		getItems = func(q string) []listing.Item {
			var items []listing.Item
			var itemsMutex sync.Mutex
			collect := func(item listing.Item) {
				itemsMutex.Lock()
				defer itemsMutex.Unlock()
				items = append(items, item)
			}
			valuesCb := func(ch <-chan interface{}) {
				for v := range ch {
					if item, itemOk := getListingItem(v); itemOk {
						collect(item)
					}
				}
			}
			bytesCb := func(r *os.File) {
				buffered := bufio.NewReader(r)
				for {
					line, err := buffered.ReadString('\n')
					if line != "" {
						s := strings.TrimSuffix(line, "\n")
						collect(listing.Item{ToAccept: s, ToShow: ui.T(s)})
					}
					if err != nil {
						break
					}
				}
			}
			err := fm.CallWithOutputCallback(
				fn, []interface{}{q}, eval.NoOpts, valuesCb, bytesCb)
			// TODO(xiaq): Report the error.
			_ = err
			return items
		}
	} else {
		getItems = func(q string) []listing.Item {
			convertedItems := []listing.Item{}
			vals.Iterate(items, func(v interface{}) bool {
				toFilter, toFilterOk := getToFilter(v)
				item, itemOk := getListingItem(v)
				if toFilterOk && itemOk && strings.Contains(toFilter, q) {
					// TODO(xiaq): Report type error when ok is false.
					convertedItems = append(convertedItems, item)
				}
				return true
			})
			return convertedItems
		}
	}

	listing.Start(app, listing.Config{
		Binding: binding,
		Caption: opts.Caption,
		GetItems: func(q string) ([]listing.Item, int) {
			items := getItems(q)
			selected := 0
			if opts.KeepBottom {
				selected = len(items) - 1
			}
			return items, selected
		},
		Accept: func(s string) bool {
			callWithNotifyPorts(app, fm.Evaler, opts.Accept, s)
			return false
		},
		AutoAccept: opts.AutoAccept,
	})
}

func getToFilter(v interface{}) (string, bool) {
	toFilterValue, _ := vals.Index(v, "to-filter")
	toFilter, toFilterOk := toFilterValue.(string)
	return toFilter, toFilterOk
}

func getListingItem(v interface{}) (item listing.Item, ok bool) {
	toAcceptValue, _ := vals.Index(v, "to-accept")
	toAccept, toAcceptOk := toAcceptValue.(string)
	toShowValue, _ := vals.Index(v, "to-show")
	toShow, toShowOk := toShowValue.(ui.Text)
	if toShowString, ok := toShowValue.(string); ok {
		toShow = ui.T(toShowString)
		toShowOk = true
	}
	return listing.Item{ToAccept: toAccept, ToShow: toShow}, toAcceptOk && toShowOk
}
