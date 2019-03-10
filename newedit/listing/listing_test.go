package listing

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/types"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

func TestModeLine(t *testing.T) {
	m := Mode{}
	m.Start(StartConfig{Name: "LISTING"})
	wantRenderer := ui.NewModeLineRenderer(" LISTING ", "")
	if renderer := m.ModeLine(); !reflect.DeepEqual(renderer, wantRenderer) {
		t.Errorf("m.ModeLine() = %v, want %v", renderer, wantRenderer)
	}
}

func TestModeRenderFlag(t *testing.T) {
	m := Mode{}
	if flag := m.ModeRenderFlag(); flag != 0 {
		t.Errorf("m.ModeRenderFlag() = %v, want 0", flag)
	}
}

func TestHandleEvent_CallsKeyHandler(t *testing.T) {
	m := Mode{}
	key := ui.Key{'a', 0}
	var calledKey ui.Key
	m.Start(StartConfig{KeyHandler: func(k ui.Key) types.HandlerAction {
		calledKey = k
		return types.CommitCode
	}})
	a := m.HandleEvent(tty.KeyEvent(key), &types.State{})
	if calledKey != key {
		t.Errorf("KeyHandler called with %v, want %v", calledKey, key)
	}
	if a != types.CommitCode {
		t.Errorf("m.HandleEvent returns %v, want CommitCode", a)
	}
}

func TestHandleEvent_DefaultHandler(t *testing.T) {
	m := Mode{}
	st := types.State{}
	st.SetMode(&m)
	m.HandleEvent(tty.KeyEvent{'[', ui.Ctrl}, &st)
	if st.Mode() != nil {
		t.Errorf("C-[ of the default handler did not set mode to nil")
	}
}

func TestHandleEvent_NonKeyEvent(t *testing.T) {
	m := Mode{}
	a := m.HandleEvent(tty.MouseEvent{}, &types.State{})
	if a != types.NoAction {
		t.Errorf("m.HandleEvent returns %v, want NoAction", a)
	}
}

type fakeItems struct{ n int }

func (it fakeItems) Len() int { return it.n }

func (it fakeItems) Show(i int) styled.Text {
	return styled.Unstyled(strconv.Itoa(i))
}

func TestList_Normal(t *testing.T) {
	m := Mode{}
	m.Start(StartConfig{ItemsGetter: func(string) Items { return fakeItems{10} }})

	m.selected = 3
	m.first = 1

	renderer := m.List(6)

	wantBase := NewStyledTextsRenderer([]styled.Text{
		styled.Unstyled("1"),
		styled.Unstyled("2"),
		styled.Transform(styled.Unstyled("3"), "inverse"),
		styled.Unstyled("4"),
		styled.Unstyled("5"),
		styled.Unstyled("6"),
	})
	wantRenderer := ui.NewRendererWithVerticalScrollbar(wantBase, 10, 1, 7)

	if !reflect.DeepEqual(renderer, wantRenderer) {
		t.Errorf("t.List() = %v, want %v", renderer, wantRenderer)
	}
}

func TestList_NoResult(t *testing.T) {
	m := Mode{}
	m.Start(StartConfig{ItemsGetter: func(string) Items { return fakeItems{0} }})

	renderer := m.List(6)
	wantRenderer := ui.NewStringRenderer("(no result)")

	if !reflect.DeepEqual(renderer, wantRenderer) {
		t.Errorf("t.List() = %v, want %v", renderer, wantRenderer)
	}
}

func TestList_Crop(t *testing.T) {
	m := Mode{}
	m.Start(StartConfig{ItemsGetter: func(string) Items {
		return SliceItems(styled.Unstyled("0a\n0b"),
			styled.Unstyled("1a\n1b"), styled.Unstyled("2a\n2b"))
	}})

	m.selected = 1
	renderer := m.List(4)

	wantBase := NewStyledTextsRenderer([]styled.Text{
		styled.Unstyled("0b"),
		styled.Transform(styled.Unstyled("1a"), "inverse"),
		styled.Transform(styled.Unstyled("1b"), "inverse"),
		styled.Unstyled("2a"),
	})
	wantRenderer := ui.NewRendererWithVerticalScrollbar(wantBase, 3, 0, 3)

	if !reflect.DeepEqual(renderer, wantRenderer) {
		t.Errorf("t.List() = %v, want %v", renderer, wantRenderer)
	}
}

var Args = tt.Args

func TestFindWindow(t *testing.T) {
	tt.Test(t, tt.Fn("findWindow", findWindow), tt.Table{
		// selected = 0: always show a widow starting from 0, regardless of
		// the value of oldFirst
		Args(fakeItems{10}, 0, 0, 6).Rets(0, 0),
		Args(fakeItems{10}, 1, 0, 6).Rets(0, 0),
		// selected = n-1: always show a window ending at n-1, regardless of the
		// value of oldFirst
		Args(fakeItems{10}, 0, 9, 6).Rets(4, 0),
		Args(fakeItems{10}, 8, 9, 6).Rets(4, 0),
		// selected = 3, oldFirst = 2 (likely because previous selected = 4).
		// Adjust first -> 1 to satisfy the upward respect distance of 2.
		Args(fakeItems{10}, 2, 3, 6).Rets(1, 0),
		// selected = 6, oldFirst = 2 (likely because previous selected = 7).
		// Adjust first -> 3 to satisfy the downward respect distance of 2.
		Args(fakeItems{10}, 2, 6, 6).Rets(3, 0),

		// There is not enough budget to achieve respect distance on both sides.
		// Split the budget in half.
		Args(fakeItems{10}, 1, 3, 3).Rets(2, 0),
		Args(fakeItems{10}, 0, 3, 3).Rets(2, 0),

		// There is just enough distance to fit the selected item. Only show the
		// selected item.
		Args(fakeItems{10}, 0, 2, 1).Rets(2, 0),
	})
}
