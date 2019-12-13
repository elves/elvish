package cliedit

import (
	"strings"

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

func listingStartCustom(app cli.App, ev *eval.Evaler, opts customListingOpts, items vals.List) {
	var binding el.Handler
	if opts.Binding.Map != nil {
		binding = newMapBinding(app, ev, vars.FromPtr(&opts.Binding))
	}
	listing.Start(app, listing.Config{
		Binding: binding,
		Caption: opts.Caption,
		GetItems: func(q string) ([]listing.Item, int) {
			parsedItems := []listing.Item{}
			vals.Iterate(items, func(v interface{}) bool {
				toFilter, item, ok := indexListingItem(v)
				if ok && strings.Contains(toFilter, q) {
					// TODO(xiaq): Report type error when ok is false.
					parsedItems = append(parsedItems, item)
				}
				return true
			})
			selected := 0
			if opts.KeepBottom {
				selected = len(parsedItems) - 1
			}
			return parsedItems, selected
		},
		Accept: func(s string) bool {
			callWithNotifyPorts(app, ev, opts.Accept, s)
			return false
		},
		AutoAccept: opts.AutoAccept,
	})
}

func indexListingItem(v interface{}) (toFilter string, item listing.Item, ok bool) {
	toFilterValue, _ := vals.Index(v, "to-filter")
	toFilter, toFilterOk := toFilterValue.(string)
	toAcceptValue, _ := vals.Index(v, "to-accept")
	toAccept, toAcceptOk := toAcceptValue.(string)
	toShowValue, _ := vals.Index(v, "to-show")
	toShow, toShowOk := toShowValue.(ui.Text)
	if toShowString, ok := toShowValue.(string); ok {
		toShow = ui.T(toShowString)
		toShowOk = true
	}
	return toFilter,
		listing.Item{ToAccept: toAccept, ToShow: toShow},
		toFilterOk && toAcceptOk && toShowOk
}
