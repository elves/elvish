package core

import "github.com/elves/elvish/edit/ui"

// An implementation of Mode. The HandleKey method returns CommitCode after a
// certain number of key events and keeps the key event history.
type fakeMode struct {
	nkeys int

	keys []ui.Key
}

func newFakeMode() *fakeMode {
	return &fakeMode{}
}

func (m *fakeMode) setMaxKeys(n int) {
	m.nkeys = n
}

func (m *fakeMode) ModeLine() ui.Renderer          { return nil }
func (m *fakeMode) ModeRenderFlag() ModeRenderFlag { return 0 }

func (m *fakeMode) HandleKey(k ui.Key, st *State) HandlerAction {
	m.keys = append(m.keys, k)
	if len(m.keys) == m.nkeys {
		return CommitCode
	}
	return 0
}
