package el

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/pkg/cli/term"
)

// RenderTest is a test case to be used in TestRenderer.
type RenderTest struct {
	Name   string
	Given  Renderer
	Width  int
	Height int
	Want   interface{ Buffer() *term.Buffer }
}

// TestRender runs the given Renderer tests.
func TestRender(t *testing.T, tests []RenderTest) {
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

// HandleTest is a test case to be used in TestHandle.
type HandleTest struct {
	Name   string
	Given  Handler
	Event  term.Event
	Events []term.Event

	WantNewState  interface{}
	WantUnhandled bool
}

// TestHandle runs the given Handler tests.
func TestHandle(t *testing.T, tests []HandleTest) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Helper()
			var handled bool
			switch {
			case test.Event != nil && test.Events != nil:
				t.Fatal("Malformed test case: both Event and Events non-nil:",
					test.Event, test.Events)
			case test.Event == nil && test.Events == nil:
				t.Fatal("Malformed test case: both Event and Events nil")
			case test.Event != nil:
				handled = test.Given.Handle(test.Event)
			default: // test.Events != nil
				for _, event := range test.Events {
					handled = test.Given.Handle(event)
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

func getState(v interface{}) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = reflect.Indirect(rv)
	}
	return rv.FieldByName("State").Interface()
}
