package core

import "github.com/elves/elvish/edit/ui"

// An implementation of Mode. The HandleKey method returns CommitCode after a
// certain number of key events and keeps the key event history.
type fakeMode struct {
	nkeys int

	keys []ui.Key
}

func newFakeMode(n int) *fakeMode {
	return &fakeMode{n, nil}
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
