package core

import "github.com/elves/elvish/edit/ui"

// An implementation of Mode. The HandleKey method returns CommitCode after a
// certain number of key events and keeps the key event history, and the
// ModeLine methods returns a predefined value.
type fakeMode struct {
	maxKeys int

	keys []ui.Key
}

func (m *fakeMode) ModeLine() ui.Renderer          { return nil }
func (m *fakeMode) ModeRenderFlag() ModeRenderFlag { return 0 }

func (m *fakeMode) HandleKey(k ui.Key, st *State) HandlerAction {
	m.keys = append(m.keys, k)
	if len(m.keys) == m.maxKeys {
		return CommitCode
	}
	return 0
}
