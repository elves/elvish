package comps

import (
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/ui"
)

// TODO: Generate these automatically.

type listItemsStructOfFunc struct {
	LenImpl  func() int          `elvish:"len"`
	GetImpl  func(i int) any     `elvish:"get"`
	ShowImpl func(i int) ui.Text `elvish:"show"`
}

var _ ListItems = listItemsStructOfFunc{}
var _ = etk.RegisterStructOfFuncForInterface[listItemsStructOfFunc, ListItems]()

func (sof listItemsStructOfFunc) Len() int           { return sof.LenImpl() }
func (sof listItemsStructOfFunc) Get(i int) any      { return sof.GetImpl(i) }
func (sof listItemsStructOfFunc) Show(i int) ui.Text { return sof.ShowImpl(i) }

type styleLinerStructOfFunc struct {
	StyleLineImpl func(i int) ui.Styling
}

var _ StyleLiner = styleLinerStructOfFunc{}
var _ = etk.RegisterStructOfFuncForInterface[styleLinerStructOfFunc, StyleLiner]()

func (sof styleLinerStructOfFunc) StyleLine(i int) ui.Styling { return sof.StyleLineImpl(i) }
