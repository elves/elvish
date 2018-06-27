package core

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
)

var (
	ka     = ui.Key{Rune: 'a'}
	kb     = ui.Key{Rune: 'b'}
	kc     = ui.Key{Rune: 'c'}
	kEnter = ui.Key{Rune: ui.Enter}

	keysABCEnter   = []ui.Key{ka, kb, kc, kEnter}
	eventsABCEnter = []tty.Event{
		tty.KeyEvent(ka), tty.KeyEvent(kb),
		tty.KeyEvent(kc), tty.KeyEvent(kEnter)}
)

func TestRead_PassesInputEventsToMode(t *testing.T) {
	ed := NewEditor(newFakeTTY(24, 80, eventsABCEnter))
	m := newFakeMode(len(eventsABCEnter))
	ed.state.Mode = m

	ed.Read()

	if !reflect.DeepEqual(m.keys, keysABCEnter) {
		t.Errorf("Mode gets keys %v, want %v", m.keys, keysABCEnter)
	}
}

func TestRead_CallsBeforeReadlineOnce(t *testing.T) {
	ed := NewEditor(newFakeTTY(24, 80, eventsABCEnter))

	called := 0
	ed.config.BeforeReadline = []func(){func() { called++ }}

	ed.Read()

	if called != 1 {
		t.Errorf("BeforeReadline hook called %d times, want 1", called)
	}
}

func TestRead_CallsAfterReadlineOnceWithCode(t *testing.T) {
	ed := NewEditor(newFakeTTY(24, 80, eventsABCEnter))

	called := 0
	code := ""
	ed.config.AfterReadline = []func(string){func(s string) {
		called++
		code = s
	}}

	ed.Read()

	if called != 1 {
		t.Errorf("AfterReadline hook called %d times, want 1", called)
	}
	if code != "abc" {
		t.Errorf("AfterReadline hook called with %q, want %q", code, "abc")
	}
}
