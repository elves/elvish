package core

import "github.com/elves/elvish/edit/ui"

// An implementation of Mode. Its HandleKey method returns CommitCode after a
// certain number of key events and keeps the key event history, and its
// ModeLine and ModeRenderFlag methods returns predefined values.
type fakeMode struct {
	maxKeys        int
	modeLine       ui.Renderer
	modeRenderFlag ModeRenderFlag

	keys []ui.Key
}

func (m *fakeMode) ModeLine() ui.Renderer { return m.modeLine }

func (m *fakeMode) ModeRenderFlag() ModeRenderFlag { return m.modeRenderFlag }

func (m *fakeMode) HandleKey(k ui.Key, st *State) HandlerAction {
	m.keys = append(m.keys, k)
	if len(m.keys) == m.maxKeys {
		return CommitCode
	}
	return 0
}
