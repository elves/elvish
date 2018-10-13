package core

import (
	"reflect"
	"testing"
)

func TestRead_ReturnsReturnValueOfHandleCb(t *testing.T) {
	handleCbRet := "lorem ipsum"
	ed := newLoop()
	ed.HandleCb(quitOn("^D", handleCbRet))
	go supplyInputs(ed, "^D")
	buf, _ := ed.Run()
	if buf != handleCbRet {
		t.Errorf("Read returns %v, want %v", buf, handleCbRet)
	}
}

func TestRead_PassesInputEventsToHandler(t *testing.T) {
	inputPassedEvents := []event{"foo", "bar", "lorem", "ipsum", "^D"}
	var handlerGotEvents []event
	handler := func(e event) (string, bool) {
		handlerGotEvents = append(handlerGotEvents, e)
		return "", e == "^D"
	}

	ed := newLoop()
	ed.HandleCb(handler)
	go supplyInputs(ed, inputPassedEvents...)

	_, _ = ed.Run()
	if !reflect.DeepEqual(handlerGotEvents, inputPassedEvents) {
		t.Errorf("Handler got events %v, expect same as events passed to input (%v)",
			handlerGotEvents, inputPassedEvents)
	}
}

func TestRead_CallsDrawWhenRedrawRequestedBeforeRead(t *testing.T) {
	testReadCallsDrawWhenRedrawRequestedBeforeRead(t, true, fullRedraw)
	testReadCallsDrawWhenRedrawRequestedBeforeRead(t, false, 0)
}

func testReadCallsDrawWhenRedrawRequestedBeforeRead(t *testing.T, full bool, wantRedrawFlag redrawFlag) {
	t.Helper()

	var gotRedrawFlag redrawFlag
	drawSeq := 0
	doneCh := make(chan struct{})
	drawer := func(full redrawFlag) {
		if drawSeq == 0 {
			gotRedrawFlag = full
			close(doneCh)
		}
		drawSeq++
	}

	ed := newLoop()
	ed.HandleCb(quitOn("^D", ""))
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
	testReadCallsDrawWhenRedrawRequestedAfterFirstDraw(t, true, fullRedraw)
	testReadCallsDrawWhenRedrawRequestedAfterFirstDraw(t, false, 0)
}

func testReadCallsDrawWhenRedrawRequestedAfterFirstDraw(t *testing.T, full bool, wantRedrawFlag redrawFlag) {
	t.Helper()

	var gotRedrawFlag redrawFlag
	drawSeq := 0
	firstDrawCalledCh := make(chan struct{})
	doneCh := make(chan struct{})
	drawer := func(flag redrawFlag) {
		if drawSeq == 0 {
			close(firstDrawCalledCh)
		} else if drawSeq == 1 {
			gotRedrawFlag = flag
			close(doneCh)
		}
		drawSeq++
	}

	ed := newLoop()
	ed.HandleCb(quitOn("^D", ""))
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

func supplyInputs(ed *loop, events ...event) {
	for _, event := range events {
		ed.Input(event)
	}
}

// Returns a HandleCb that quits on a trigger event.
func quitOn(retTrigger event, ret string) handleCb {
	return func(e event) (string, bool) {
		return ret, e == retTrigger
	}
}
