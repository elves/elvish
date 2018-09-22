package core

import (
	"fmt"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/types"
)

// An implementation of Mode. Its HandleKey method returns CommitCode after a
// certain number of key events and keeps the key event history, and its
// ModeLine and ModeRenderFlag methods returns predefined values.
type fakeMode struct {
	maxKeys        int
	modeLine       ui.Renderer
	modeRenderFlag types.ModeRenderFlag

	keysHandled []ui.Key
}

func (m *fakeMode) ModeLine() ui.Renderer {
	return m.modeLine
}

func (m *fakeMode) ModeRenderFlag() types.ModeRenderFlag { return m.modeRenderFlag }

func (m *fakeMode) HandleKey(k ui.Key, _ *types.State) types.HandlerAction {
	m.keysHandled = append(m.keysHandled, k)
	if len(m.keysHandled) == m.maxKeys {
		return types.CommitCode
	}
	return 0
}

// A listing mode with a predefined listing.
type fakeListingMode struct {
	fakeMode
	list []string
}

func (m *fakeListingMode) List(h int) ui.Renderer {
	if h >= len(m.list) {
		return &linesRenderer{m.list}
	}
	return &linesRenderer{m.list[:h]}
}

// A listing mode whose Modeline shows how many time it is called.
type fakeListingModeWithModeline struct {
	fakeMode
	modeLineCalled int
}

func (m *fakeListingModeWithModeline) ModeLine() ui.Renderer {
	m.modeLineCalled++
	return &linesRenderer{[]string{fmt.Sprintf("#%d", m.modeLineCalled)}}
}

func (m *fakeListingModeWithModeline) List(h int) ui.Renderer {
	return nil
}
