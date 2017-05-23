package edit

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

type provider struct {
	elems    []string
	accepted int
}

func (p provider) Len() int                 { return len(p.elems) }
func (p provider) Filter(string) int        { return 0 }
func (p provider) Accept(i int, ed *Editor) { p.accepted = i }
func (p provider) ModeTitle(i int) string   { return fmt.Sprintf("test %d", i) }

func (p provider) Show(i int) (string, styled) {
	return strconv.Itoa(i), unstyled(p.elems[i])
}

var (
	mode = ModeType(233)
	p    = provider{[]string{"foo", "bar", "foobar", "lorem", "ipsum"}, -1}
	ls   = newListing(mode, p)
)

func TestListing(t *testing.T) {
	if m := ls.Mode(); m != mode {
		t.Errorf("ls.Mode() = %v, want %v", m, mode)
	}

	wantedModeLine := modeLineRenderer{"test 0", ""}
	if modeLine := ls.ModeLine(); modeLine != wantedModeLine {
		t.Errorf("ls.ModeLine() = %v, want %v", modeLine, wantedModeLine)
	}

	// Selecting the first element and rendering with height=2. We expect to see
	// the first 2 elements, with the first being shown as selected.
	testListingList(t, 0, 2, listingWithScrollBarRenderer{
		listingRenderer: listingRenderer{[]styled{
			styled{"0 foo", styleForSelected},
			styled{"1 bar", styles{}},
		}},
		n: 5, low: 0, high: 2, height: 2,
	})
	// Selecting the last element and rendering with height=2. We expect to see
	// the last 2 elements, with the last being shown as selected.
	testListingList(t, 4, 2, listingWithScrollBarRenderer{
		listingRenderer: listingRenderer{[]styled{
			styled{"3 lorem", styles{}},
			styled{"4 ipsum", styleForSelected},
		}},
		n: 5, low: 3, high: 5, height: 2,
	})
	// Selecting the middle element and rendering with height=3. We expect to
	// see the middle element and two elements around it, with the middle being
	// shown as selected.
	testListingList(t, 2, 3, listingWithScrollBarRenderer{
		listingRenderer: listingRenderer{[]styled{
			styled{"1 bar", styles{}},
			styled{"2 foobar", styleForSelected},
			styled{"3 lorem", styles{}},
		}},
		n: 5, low: 1, high: 4, height: 3,
	})
}

func testListingList(t *testing.T, i, h int, want renderer) {
	ls.selected = i
	if r := ls.List(h); !reflect.DeepEqual(r, want) {
		t.Errorf("selecting %d, ls.List(%d) = %v, want %v", i, h, r, want)
	}
}
