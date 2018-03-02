package edit

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
)

type provider struct {
	elems    []string
	accepted int
}

func (p provider) Len() int                       { return len(p.elems) }
func (p provider) Filter(string) int              { return 0 }
func (p provider) Accept(i int, ed eddefs.Editor) { p.accepted = i }
func (p provider) ModeTitle(i int) string         { return fmt.Sprintf("test %d", i) }

func (p provider) Show(i int) (string, ui.Styled) {
	return strconv.Itoa(i), ui.Unstyled(p.elems[i])
}

var (
	p  = provider{[]string{"foo", "bar", "foobar", "lorem", "ipsum"}, -1}
	ls = newListing(emptyBindingMap, p)
)

func TestListing(t *testing.T) {
	wantedModeLine := ui.NewModeLineRenderer("test 0", "")
	if modeLine := ls.ModeLine(); modeLine != wantedModeLine {
		t.Errorf("ls.ModeLine() = %v, want %v", modeLine, wantedModeLine)
	}

	// Selecting the first element and rendering with height=2. We expect to see
	// the first 2 elements, with the first being shown as selected.
	testListingList(t, 0, 2, listingWithScrollBarRenderer{
		listingRenderer: listingRenderer{[]ui.Styled{
			{"0 foo", styleForSelected},
			{"1 bar", ui.Styles{}},
		}},
		n: 5, low: 0, high: 2, height: 2,
	})
	// Selecting the last element and rendering with height=2. We expect to see
	// the last 2 elements, with the last being shown as selected.
	testListingList(t, 4, 2, listingWithScrollBarRenderer{
		listingRenderer: listingRenderer{[]ui.Styled{
			{"3 lorem", ui.Styles{}},
			{"4 ipsum", styleForSelected},
		}},
		n: 5, low: 3, high: 5, height: 2,
	})
	// Selecting the middle element and rendering with height=3. We expect to
	// see the middle element and two elements around it, with the middle being
	// shown as selected.
	testListingList(t, 2, 3, listingWithScrollBarRenderer{
		listingRenderer: listingRenderer{[]ui.Styled{
			{"1 bar", ui.Styles{}},
			{"2 foobar", styleForSelected},
			{"3 lorem", ui.Styles{}},
		}},
		n: 5, low: 1, high: 4, height: 3,
	})
}

func testListingList(t *testing.T, i, h int, want ui.Renderer) {
	ls.selected = i
	if r := ls.List(h); !reflect.DeepEqual(r, want) {
		t.Errorf("selecting %d, ls.List(%d) = %v, want %v", i, h, r, want)
	}
}
