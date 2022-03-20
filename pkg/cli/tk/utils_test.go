package tk

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/ui"
)

// renderTest is a test case to be used in TestRenderer.
type renderTest struct {
	Name   string
	Given  Renderer
	Width  int
	Height int
	Want   interface{ Buffer() *term.Buffer }
}

// testRender runs the given Renderer tests.
func testRender(t *testing.T, tests []renderTest) {
	t.Helper()
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Helper()
			buf := test.Given.Render(test.Width, test.Height)
			wantBuf := test.Want.Buffer()
			if !reflect.DeepEqual(buf, wantBuf) {
				t.Errorf("Buffer mismatch")
				t.Logf("Got: %s", buf.TTYString())
				t.Logf("Want: %s", wantBuf.TTYString())
			}
		})
	}
}

// handleTest is a test case to be used in testHandle.
type handleTest struct {
	Name   string
	Given  Handler
	Event  term.Event
	Events []term.Event

	WantNewState  any
	WantUnhandled bool
}

// testHandle runs the given Handler tests.
func testHandle(t *testing.T, tests []handleTest) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Helper()

			handler := test.Given
			oldState := getState(handler)
			defer setState(handler, oldState)

			var handled bool
			switch {
			case test.Event != nil && test.Events != nil:
				t.Fatal("Malformed test case: both Event and Events non-nil:",
					test.Event, test.Events)
			case test.Event == nil && test.Events == nil:
				t.Fatal("Malformed test case: both Event and Events nil")
			case test.Event != nil:
				handled = handler.Handle(test.Event)
			default: // test.Events != nil
				for _, event := range test.Events {
					handled = handler.Handle(event)
				}
			}
			if handled != !test.WantUnhandled {
				t.Errorf("Got handled %v, want %v", handled, !test.WantUnhandled)
			}
			if test.WantNewState != nil {
				state := getState(test.Given)
				if !reflect.DeepEqual(state, test.WantNewState) {
					t.Errorf("Got state %v, want %v", state, test.WantNewState)
				}
			}
		})
	}
}

func getState(v any) any {
	return reflectState(v).Interface()
}

func setState(v, state any) {
	reflectState(v).Set(reflect.ValueOf(state))
}

func reflectState(v any) reflect.Value {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = reflect.Indirect(rv)
	}
	return rv.FieldByName("State")
}

// Test for the test utilities.

func TestTestRender(t *testing.T) {
	testRender(t, []renderTest{
		{
			Name:  "test",
			Given: &testWidget{text: ui.T("test")},
			Width: 10, Height: 10,

			Want: term.NewBufferBuilder(10).Write("test"),
		},
	})
}

type testHandlerWithState struct {
	State testHandlerState
}

type testHandlerState struct {
	last  term.Event
	total int
}

func (h *testHandlerWithState) Handle(e term.Event) bool {
	if e == term.K('x') {
		return false
	}
	h.State.last = e
	h.State.total++
	return true
}

func TestTestHandle(t *testing.T) {
	testHandle(t, []handleTest{
		{
			Name:  "WantNewState",
			Given: &testHandlerWithState{},
			Event: term.K('a'),

			WantNewState: testHandlerState{last: term.K('a'), total: 1},
		},
		{
			Name:   "Multiple events",
			Given:  &testHandlerWithState{},
			Events: []term.Event{term.K('a'), term.K('b')},

			WantNewState: testHandlerState{last: term.K('b'), total: 2},
		},
		{
			Name:  "WantUnhaneld",
			Given: &testHandlerWithState{},
			Event: term.K('x'),

			WantUnhandled: true,
		},
	})
}
