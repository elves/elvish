package abstract

import (
	"errors"
	"reflect"
	"testing"
)

func TestRead_returnsReturnValueOfHandler(t *testing.T) {
	ed := NewEditor(inputOf("^D"), quitOn("^D", "buffer"))
	buf, _ := ed.Read()
	if buf != "buffer" {
		t.Errorf("Read returns %v, want buffer", buf)
	}
}

func TestRead_callsSetupAndRestore(t *testing.T) {
	ed := NewEditor(inputOf("^D"), quitOn("^D", ""))

	setupCalled, restoreCalled := 0, 0
	ed.SetupCb(func() (func(), error) {
		setupCalled++
		return func() { restoreCalled++ }, nil
	})

	_, _ = ed.Read()
	if setupCalled != 1 {
		t.Errorf("setup called %d times, expect once", setupCalled)
	}
	if restoreCalled != 1 {
		t.Errorf("restore called %d times, expect once", restoreCalled)
	}
}

func TestRead_returnsErrorFromSetuper(t *testing.T) {
	ed := NewEditor(inputOf("^D"), quitOn("^D", ""))
	ed.SetupCb(badSetuper)

	_, err := ed.Read()
	if err != errBadSetuper {
		t.Errorf("Read returned with error %v, expect errBadSetuper", err)
	}
}

func TestRead_doesntCallInputWhenSetupErrors(t *testing.T) {
	inputCalled := false
	input := func() (<-chan Event, func()) {
		inputCalled = true
		return nil, func() {}
	}
	handler := func(Event) (string, bool) { return "", false }

	ed := NewEditor(input, handler)

	ed.SetupCb(badSetuper)

	_, _ = ed.Read()
	if inputCalled {
		t.Errorf("Input still called when setup returned error")
	}
}

func TestRead_passInputEventsToHandler(t *testing.T) {
	inputPassedEvents := []Event{"foo", "bar", "lorem", "ipsum", "^D"}
	input := inputOf(inputPassedEvents...)
	var handlerGotEvents []Event
	handler := func(e Event) (string, bool) {
		handlerGotEvents = append(handlerGotEvents, e)
		return "", e == "^D"
	}

	ed := NewEditor(input, handler)

	_, _ = ed.Read()
	if !reflect.DeepEqual(handlerGotEvents, inputPassedEvents) {
		t.Errorf("Handler got events %v, expect same as events passed to input (%v)",
			handlerGotEvents, inputPassedEvents)
	}
}

func TestRead_callsDrawWhenRedrawRequestedBeforeRead(t *testing.T) {
	testRead_callsDrawWhenRedrawRequestedBeforeRead(t, true)
	testRead_callsDrawWhenRedrawRequestedBeforeRead(t, false)
}

func testRead_callsDrawWhenRedrawRequestedBeforeRead(t *testing.T, arg bool) {
	var drawerGotArg RedrawFlag
	drawSeq := 0
	doneCh := make(chan struct{})
	drawer := func(arg RedrawFlag) {
		if drawSeq == 0 {
			drawerGotArg = arg
			close(doneCh)
		}
		drawSeq++
	}

	ed := NewEditor(inputAfter(doneCh, "^D"), quitOn("^D", ""))
	ed.RedrawCb(drawer)
	ed.Redraw(arg)
	_, _ = ed.Read()
	if drawerGotArg != FullRedraw {
		t.Errorf("Drawer got args %v, expect FullRedraw", drawerGotArg)
	}
}

func TestRead_callsDrawWhenRedrawRequestedAfterFirstDraw(t *testing.T) {
	testRead_callsDrawWhenRedrawRequestedAfterFirstDraw(t, true, FullRedraw)
	testRead_callsDrawWhenRedrawRequestedAfterFirstDraw(t, false, 0)
}

func testRead_callsDrawWhenRedrawRequestedAfterFirstDraw(t *testing.T, full bool, wantRedrawrFlag RedrawFlag) {
	var gotRedrawFlag RedrawFlag
	drawSeq := 0
	firstDrawCalledCh := make(chan struct{})
	doneCh := make(chan struct{})
	drawer := func(flag RedrawFlag) {
		if drawSeq == 0 {
			close(firstDrawCalledCh)
		} else if drawSeq == 1 {
			gotRedrawFlag = flag
			close(doneCh)
		}
		drawSeq++
	}

	ed := NewEditor(inputAfter(doneCh, "^D"), quitOn("^D", ""))
	ed.RedrawCb(drawer)
	go func() {
		<-firstDrawCalledCh
		ed.Redraw(full)
	}()
	_, _ = ed.Read()
	if gotRedrawFlag != wantRedrawrFlag {
		t.Errorf("Drawer got args %v, want %v instead",
			gotRedrawFlag, wantRedrawrFlag)
	}
}

// Helpers.

var errBadSetuper = errors.New("418 I'm a teapot")

func badSetuper() (func(), error) { return nil, errBadSetuper }

func inputOf(events ...Event) InputCb {
	eventCh := make(chan Event, len(events))
	for _, event := range events {
		eventCh <- event
	}
	return func() (<-chan Event, func()) {
		return eventCh, func() {}
	}
}

func inputAfter(ch <-chan struct{}, events ...Event) InputCb {
	eventCh := make(chan Event)
	go func() {
		<-ch
		for _, event := range events {
			eventCh <- event
		}
	}()
	return func() (<-chan Event, func()) {
		return eventCh, func() {}
	}
}

// Returns a HandleCb that quits on a trigger event.
func quitOn(retTrigger Event, ret string) HandleCb {
	return func(e Event) (string, bool) {
		return ret, e == retTrigger
	}
}
