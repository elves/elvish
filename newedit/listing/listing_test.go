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

// Implementation of Items that emulates a list of numbers from 0 to n-1.
type fakeItems struct{ n int }

func (it fakeItems) Len() int { return it.n }

func (it fakeItems) Show(i int) styled.Text {
	return styled.Unstyled(strconv.Itoa(i))
}

func (it fakeItems) Accept(int, *types.State) {}

// Implementation of Items that emulate 10 empty texts, but can be accepted.
type fakeAcceptableItems struct{ accept func(int, *types.State) }

func (it fakeAcceptableItems) Len() int { return 10 }

func (it fakeAcceptableItems) Show(int) styled.Text {
	return styled.Unstyled("")
}

func (it fakeAcceptableItems) Accept(i int, st *types.State) {
	it.accept(i, st)
}

func TestModeLine(t *testing.T) {
	m := Mode{}
	m.Start(StartConfig{Name: "LISTING"})
	m.state.filter = "filter"
	wantRenderer := ui.NewModeLineRenderer(" LISTING ", "filter")
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

func TestHandleEvent_DefaultBinding(t *testing.T) {
	m := Mode{}
	m.Start(StartConfig{ItemsGetter: func(string) Items {
		return fakeItems{10}
	}})
	st := types.State{}
	st.SetMode(&m)

	m.HandleEvent(tty.KeyEvent{ui.Down, 0}, &st)
	if m.state.selected != 1 {
		t.Errorf("Down did not move selection down")
	}

	m.HandleEvent(tty.KeyEvent{ui.Up, 0}, &st)
	if m.state.selected != 0 {
		t.Errorf("Up did not move selection up")
	}

	m.HandleEvent(tty.KeyEvent{ui.Up, 0}, &st)
	if m.state.selected != 0 {
		t.Errorf("Up did not stop at first item")
	}

	m.HandleEvent(tty.KeyEvent{ui.Tab, ui.Shift}, &st)
	if m.state.selected != 9 {
		t.Errorf("Shift-Tab did not wrap to last item")
	}

	m.HandleEvent(tty.KeyEvent{ui.Tab, 0}, &st)
	if m.state.selected != 0 {
		t.Errorf("Tab did not wrap to first item")
	}

	m.HandleEvent(tty.KeyEvent{ui.Tab, 0}, &st)
	if m.state.selected != 1 {
		t.Errorf("Tab did not move selection down")
	}

	m.HandleEvent(tty.KeyEvent{ui.Tab, ui.Shift}, &st)
	if m.state.selected != 0 {
		t.Errorf("Shift-Tab did not move selection up")
	}

	m.HandleEvent(tty.KeyEvent{'F', ui.Ctrl}, &st)
	if !m.state.filtering {
		t.Errorf("Ctrl-F does not enable filtering")
	}

	m.HandleEvent(tty.KeyEvent{'[', ui.Ctrl}, &st)
	if st.Mode() != nil {
		t.Errorf("Ctrl-[ did not set mode to nil")
	}
}

func TestDefaultHandler_Filtering(t *testing.T) {
	m := Mode{}
	filter := ""
	m.Start(StartConfig{ItemsGetter: func(f string) Items {
		filter = f
		return fakeItems{10}
	}})
	m.state.filtering = true
	st := types.State{}
	st.SetMode(&m)

	st.SetBindingKey(ui.K('a'))
	m.DefaultHandler(&st)
	if m.state.filter != "a" {
		t.Errorf("Printable key did not append to filter")
	}
	if filter != "a" {
		t.Errorf("Filter in state is %q, not updated", filter)
	}

	m.state.filter = "hello world"
	st.SetBindingKey(ui.K(ui.Backspace))
	m.DefaultHandler(&st)
	if m.state.filter != "hello worl" {
		t.Errorf("Backspace did not remove last char of filter")
	}
	if filter != "hello worl" {
		t.Errorf("Filter in state is %q, not updated", filter)
	}

	st.SetBindingKey(ui.K('A', ui.Ctrl))
	m.DefaultHandler(&st)
	wantNotes := []string{"Unbound: Ctrl-A"}
	if !reflect.DeepEqual(st.Raw.Notes, wantNotes) {
		t.Errorf("Unbound key made notes %v, want %v", st.Raw.Notes, wantNotes)
	}
}

func TestDefaultHandler_NotFiltering(t *testing.T) {
	m := Mode{}
	m.Start(StartConfig{ItemsGetter: func(f string) Items {
		return fakeItems{10}
	}})
	st := types.State{}
	st.SetMode(&m)

	st.SetBindingKey(ui.K('a'))
	m.DefaultHandler(&st)
	if st.Mode() != nil {
		t.Errorf("Mode not reset")
	}
}

func TestDefaultHandler_AutoAccept(t *testing.T) {
	m := Mode{}
	m.Start(StartConfig{
		ItemsGetter: func(f string) Items {
			if f == "xy" {
				return fakeItems{1}
			}
			return fakeItems{10}
		},
		StartFilter: true,
		AutoAccept:  true,
	})

	st := types.State{}
	st.SetMode(&m)

	st.SetBindingKey(ui.K('x'))
	m.DefaultHandler(&st)
	if st.Mode() == nil {
		t.Errorf("Auto-accepted too early")
	}

	st.SetBindingKey(ui.K('y'))
	m.DefaultHandler(&st)
	if st.Mode() != nil {
		t.Errorf("Did not auto-accept when there is only one item")
	}
}

func TestHandleEvent_NonKeyEvent(t *testing.T) {
	m := Mode{}
	a := m.HandleEvent(tty.MouseEvent{}, &types.State{})
	if a != types.NoAction {
		t.Errorf("m.HandleEvent returns %v, want NoAction", a)
	}
}

func TestMutateState(t *testing.T) {
	m := Mode{}
	m.MutateStates(func(st *State) {
		st.selected = 10
	})
	if m.state.selected != 10 {
		t.Errorf("state not mutated")
	}
}

func TestAcceptItem(t *testing.T) {
	m := Mode{}
	accepted := -1
	m.Start(StartConfig{ItemsGetter: func(string) Items {
		return fakeAcceptableItems{func(i int, st *types.State) { accepted = i }}
	}})
	m.state.selected = 7
	m.AcceptItem(&types.State{})
	if accepted != 7 {
		t.Errorf("accept called with %v, want 7", accepted)
	}
}

func TestAcceptItemAndClose(t *testing.T) {
	m := Mode{}
	accepted := -1
	m.Start(StartConfig{ItemsGetter: func(string) Items {
		return fakeAcceptableItems{func(i int, st *types.State) { accepted = i }}
	}})
	m.state.selected = 7
	st := &types.State{}
	st.SetMode(&m)
	m.AcceptItemAndClose(st)
	if accepted != 7 {
		t.Errorf("accept called with %v, want 7", accepted)
	}
	if st.Raw.Mode != nil {
		t.Errorf("mode not reset")
	}
}

func TestList_Normal(t *testing.T) {
	m := Mode{}
	m.Start(StartConfig{ItemsGetter: func(string) Items { return fakeItems{10} }})

	m.state.selected = 3
	m.state.first = 1

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

	m.state.selected = 1
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
