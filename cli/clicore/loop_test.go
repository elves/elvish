package clicore

import (
	"fmt"
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
	handler := func(e event) handleResult {
		handlerGotEvents = append(handlerGotEvents, e)
		return handleResult{quit: e == "^D"}
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
	return func(e event) handleResult {
		return handleResult{quit: e == retTrigger, buffer: ret}
	}
}

func ExampleLoop() {
	buffer := ""
	firstDrawerCall := true
	drawer := func(flag redrawFlag) {
		// Because the consumption of events is batched, calls to the drawer is
		// nondeterministic except for the first and final calls.
		switch {
		case firstDrawerCall:
			fmt.Printf("initial buffer is %q\n", buffer)
			firstDrawerCall = false
		case flag&finalRedraw != 0:
			fmt.Printf("final buffer is %q\n", buffer)
		}
	}
	handler := func(ev event) handleResult {
		if ev == '\n' {
			return handleResult{quit: true, buffer: buffer}
		}
		buffer += string(ev.(rune))
		return handleResult{}
	}

	ed := newLoop()
	ed.HandleCb(handler)
	go func() {
		for _, event := range "echo\n" {
			ed.Input(event)
		}
	}()
	ed.RedrawCb(drawer)
	buf, err := ed.Run()
	fmt.Printf("returned buffer is %q\n", buf)
	fmt.Printf("returned error is %v\n", err)
	// Output:
	// initial buffer is ""
	// final buffer is "echo"
	// returned buffer is "echo"
	// returned error is <nil>
}
