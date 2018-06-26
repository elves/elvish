package core

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
)

var (
	ka = ui.Key{Rune: 'a'}
	kb = ui.Key{Rune: 'b'}
	kc = ui.Key{Rune: 'c'}
)

func TestRead_PassesInputEventsToMode(t *testing.T) {
	r := newFakeReader(tty.KeyEvent(ka), tty.KeyEvent(kb), tty.KeyEvent(kc))
	w := newFakeWriter()
	tm := newFakeTTY(24, 80)

	ed := NewEditor(r, w, tm)

	m := newFakeMode(3)
	ed.state.Mode = m

	_, err := ed.Read()

	if err != nil {
		t.Errorf("Read returns error %v", err)
	}

	wantKeys := []ui.Key{ka, kb, kc}
	if !reflect.DeepEqual(m.keys, wantKeys) {
		t.Errorf("Mode gets keys %v, want %v", m.keys, wantKeys)
	}
}
