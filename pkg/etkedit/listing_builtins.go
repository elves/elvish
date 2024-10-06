package edit

import (
	"errors"
	"strconv"

	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
)

func listingUp(selected, n, height int) int        { return selected - 1 }
func listingDown(selected, n, height int) int      { return selected + 1 }
func listingUpCycle(selected, n, height int) int   { return (selected - 1 + n) % n }
func listingDownCycle(selected, n, height int) int { return (selected + 1) % n }
func listingLeft(selected, n, height int) int      { return selected - height }
func listingRight(selected, n, height int) int     { return selected + height }

var errNoListingIsActive = errors.New("no listing is active")

func wrapListingSelect(ed *Editor, f func(selected, n, height int) int) func() error {
	return func() error {
		c, err := etkCtx(ed)
		if err != nil {
			return err
		}
		focus := etk.BindState(c, "focus", 0).Get()
		if focus == 0 {
			return errNoListingIsActive
		}
		addonID := strconv.Itoa(etk.BindState(c, "addons", addons{}).Get().Addons[focus-1].ID)
		items, ok := c.Get(addonID + "/list/items").(comps.ListItems)
		if !ok {
			return errNoListingIsActive
		}
		selectedVar := etk.BindState(c, addonID+"/list/selected", 0)
		contentHeight := etk.BindState(c, addonID+"/list/content-height", 0).Get()
		selectedVar.Swap(func(selected int) int {
			n := items.Len()
			newSelected := f(selected, n, contentHeight)
			if 0 <= newSelected && newSelected < n {
				return newSelected
			}
			return selected
		})
		return nil
	}
}
