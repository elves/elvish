package core

import (
	"testing"
	"time"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
)

var basicModeTests = []struct {
	events   []tty.Event
	wantCode string
	wantDot  int
}{
	// ASCII characters
	{[]tty.Event{tty.KeyEvent{Rune: 'a'}}, "a", 1},
	// Unicode characters
	{[]tty.Event{tty.KeyEvent{Rune: '代'}, tty.KeyEvent{Rune: '码'}},
		"代码", 6},
	// Backspace
	{[]tty.Event{
		tty.KeyEvent{Rune: '代'}, tty.KeyEvent{Rune: '码'},
		tty.KeyEvent{Rune: ui.Backspace}},
		"代", 3},
	// Left
	{[]tty.Event{
		tty.KeyEvent{Rune: '代'}, tty.KeyEvent{Rune: '码'},
		tty.KeyEvent{Rune: ui.Left}},
		"代码", 3},
	{[]tty.Event{
		tty.KeyEvent{Rune: '代'}, tty.KeyEvent{Rune: '码'},
		tty.KeyEvent{Rune: ui.Left}, tty.KeyEvent{Rune: ui.Left}},
		"代码", 0},
	// Right
	{[]tty.Event{
		tty.KeyEvent{Rune: '代'}, tty.KeyEvent{Rune: '码'},
		tty.KeyEvent{Rune: ui.Left}, tty.KeyEvent{Rune: ui.Left},
		tty.KeyEvent{Rune: ui.Right}},
		"代码", 3},
}

var stateUpdateTimeout = 1 * time.Second

func TestBasicMode(t *testing.T) {
	for _, test := range basicModeTests {
		terminal := newFakeTTY(test.events)
		ed := NewEditor(terminal, nil)
		go ed.ReadCode()
	checkState:
		for {
			select {
			case <-terminal.bufCh:
				if ed.state.Code == test.wantCode && ed.state.Dot == test.wantDot {
					break checkState
				}
			case <-time.After(time.Second):
				t.Errorf("Timeout waiting for matching state")
				break checkState
			}
		}
		terminal.eventCh <- tty.KeyEvent{Rune: ui.Enter}
	}
}
