package edit

import (
	"fmt"
	"strconv"
	"testing"
)

type provider struct {
	len      int
	accepted int
}

func (p provider) Len() int                 { return p.len }
func (p provider) Filter(string) int        { return 0 }
func (p provider) Accept(i int, ed *Editor) { p.accepted = i }
func (p provider) ModeTitle(i int) string   { return fmt.Sprintf("test %d", i) }

func (p provider) Show(i int) (string, styled) {
	s := strconv.Itoa(i)
	return s, styled{s, stylesFromString("1")}
}

var (
	mode = ModeType(233)
	p    = provider{10, -1}
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
}
