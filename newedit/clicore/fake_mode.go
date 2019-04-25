package clicore

import (
	"fmt"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/types"
)

// A Mode implementation useful in tests.
type fakeMode struct {
	maxKeys        int
	modeLine       ui.Renderer
	modeRenderFlag types.ModeRenderFlag

	// History of all keys HandleEvent has seen.
	keysHandled []ui.Key
}

// ModeLine returns the predefined value.
func (m *fakeMode) ModeLine() ui.Renderer {
	return m.modeLine
}

// ModeRenderFlag returns the predefined value.
func (m *fakeMode) ModeRenderFlag() types.ModeRenderFlag {
	return m.modeRenderFlag
}

// HandleEvent records all keys it has seen, and returns CommitCode after seeing
// a predefined number of keys. It ignores other events.
func (m *fakeMode) HandleEvent(e tty.Event, _ *types.State) types.HandlerAction {
	if keyEvent, ok := e.(tty.KeyEvent); ok {
		m.keysHandled = append(m.keysHandled, ui.Key(keyEvent))
		if len(m.keysHandled) == m.maxKeys {
			return types.CommitCode
		}
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
