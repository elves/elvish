package loop

import (
	"reflect"
	"testing"
)

func TestRead_returnsReturnValueOfHandleCb(t *testing.T) {
	handleCbRet := "lorem ipsum"
	ed := New(quitOn("^D", handleCbRet))
	go supplyInputs(ed, "^D")
	buf, _ := ed.Run()
	if buf != handleCbRet {
		t.Errorf("Read returns %v, want %v", buf, handleCbRet)
	}
}

func TestRead_passInputEventsToHandler(t *testing.T) {
	inputPassedEvents := []Event{"foo", "bar", "lorem", "ipsum", "^D"}
	var handlerGotEvents []Event
	handler := func(e Event) (string, bool) {
		handlerGotEvents = append(handlerGotEvents, e)
		return "", e == "^D"
	}

	ed := New(handler)
	go supplyInputs(ed, inputPassedEvents...)

	_, _ = ed.Run()
	if !reflect.DeepEqual(handlerGotEvents, inputPassedEvents) {
		t.Errorf("Handler got events %v, expect same as events passed to input (%v)",
			handlerGotEvents, inputPassedEvents)
	}
}

func TestRead_callsDrawWhenRedrawRequestedBeforeRead(t *testing.T) {
	testRead_callsDrawWhenRedrawRequestedBeforeRead(t, true, FullRedraw)
	testRead_callsDrawWhenRedrawRequestedBeforeRead(t, false, 0)
}

func testRead_callsDrawWhenRedrawRequestedBeforeRead(t *testing.T, full bool, wantRedrawFlag RedrawFlag) {
	t.Helper()

	var gotRedrawFlag RedrawFlag
	drawSeq := 0
	doneCh := make(chan struct{})
	drawer := func(full RedrawFlag) {
		if drawSeq == 0 {
			gotRedrawFlag = full
			close(doneCh)
		}
		drawSeq++
	}

	ed := New(quitOn("^D", ""))
	go func() {
		<-doneCh
		ed.Input("^D")
	}()
	ed.RedrawCb(drawer)
	ed.Redraw(full)
	_, _ = ed.Run()
	if gotRedrawFlag != wantRedrawFlag {
		t.Errorf("Drawer got flag %v, want %v", gotRedrawFlag, wantRedrawFlag)
	}
}

func TestRead_callsDrawWhenRedrawRequestedAfterFirstDraw(t *testing.T) {
	testRead_callsDrawWhenRedrawRequestedAfterFirstDraw(t, true, FullRedraw)
	testRead_callsDrawWhenRedrawRequestedAfterFirstDraw(t, false, 0)
}

func testRead_callsDrawWhenRedrawRequestedAfterFirstDraw(t *testing.T, full bool, wantRedrawFlag RedrawFlag) {
	t.Helper()

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

	ed := New(quitOn("^D", ""))
	go func() {
		<-doneCh
		ed.Input("^D")
	}()
	ed.RedrawCb(drawer)
	go func() {
		<-firstDrawCalledCh
		ed.Redraw(full)
	}()
	_, _ = ed.Run()
	if gotRedrawFlag != wantRedrawFlag {
		t.Errorf("Drawer got flag %v, want %v", gotRedrawFlag, wantRedrawFlag)
	}
}

// Helpers.

func supplyInputs(ed *Loop, events ...Event) {
	for _, event := range events {
		ed.Input(event)
	}
}

// Returns a HandleCb that quits on a trigger event.
func quitOn(retTrigger Event, ret string) HandleCb {
	return func(e Event) (string, bool) {
		return ret, e == retTrigger
	}
}
