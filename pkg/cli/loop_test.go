package cli

import (
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestRead_PassesInputEventsToHandler(t *testing.T) {
	var handlerGotEvents []event

	lp := newLoop()
	lp.HandleCb(func(e event) {
		handlerGotEvents = append(handlerGotEvents, e)
		if e == "^D" {
			lp.Return("", nil)
		}
	})

	inputPassedEvents := []event{"foo", "bar", "lorem", "ipsum", "^D"}
	supplyInputs(lp, inputPassedEvents...)

	_, _ = lp.Run()
	if !reflect.DeepEqual(handlerGotEvents, inputPassedEvents) {
		t.Errorf("Handler got events %v, expect same as events passed to input (%v)",
			handlerGotEvents, inputPassedEvents)
	}
}

func TestLoop_RunReturnsAfterReturnCalled(t *testing.T) {
	lp := newLoop()
	lp.HandleCb(func(event) { lp.Return("buffer", io.EOF) })
	supplyInputs(lp, "x")
	buf, err := lp.Run()
	if buf != "buffer" || err != io.EOF {
		fmt.Printf("Run -> (%v, %v), want (%v, %v)", buf, err, "buffer", io.EOF)
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

	lp := newLoop()
	lp.HandleCb(quitOn(lp, "^D", "", nil))
	go func() {
		<-doneCh
		lp.Input("^D")
	}()
	lp.RedrawCb(drawer)
	lp.Redraw(full)
	_, _ = lp.Run()
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

	lp := newLoop()
	lp.HandleCb(quitOn(lp, "^D", "", nil))
	go func() {
		<-doneCh
		lp.Input("^D")
	}()
	lp.RedrawCb(drawer)
	go func() {
		<-firstDrawCalledCh
		lp.Redraw(full)
	}()
	_, _ = lp.Run()
	if gotRedrawFlag != wantRedrawFlag {
		t.Errorf("Drawer got flag %v, want %v", gotRedrawFlag, wantRedrawFlag)
	}
}

// Helpers.

func supplyInputs(lp *loop, events ...event) {
	for _, event := range events {
		lp.Input(event)
	}
}

// Returns a HandleCb that quits on a trigger event.
func quitOn(lp *loop, retTrigger event, ret string, err error) handleCb {
	return func(e event) {
		if e == retTrigger {
			lp.Return(ret, err)
		}
	}
}

func TestLoop_FullLifecycle(t *testing.T) {
	// A test for the entire lifecycle of a loop.

	var initialBuffer, finalBuffer string

	buffer := ""
	firstDrawerCall := true
	drawer := func(flag redrawFlag) {
		// Because the consumption of events is batched, calls to the drawer is
		// nondeterministic except for the first and final calls.
		switch {
		case firstDrawerCall:
			initialBuffer = buffer
			firstDrawerCall = false
		case flag&finalRedraw != 0:
			finalBuffer = buffer
		}
	}

	lp := newLoop()
	lp.HandleCb(func(e event) {
		if e == '\n' {
			lp.Return(buffer, nil)
			return
		}
		buffer += string(e.(rune))
	})
	go func() {
		for _, event := range "echo\n" {
			lp.Input(event)
		}
	}()
	lp.RedrawCb(drawer)
	returnedBuffer, err := lp.Run()

	if initialBuffer != "" {
		t.Errorf("got initial buffer %q, want %q", initialBuffer, "")
	}
	if finalBuffer != "echo" {
		t.Errorf("got final buffer %q, want %q", finalBuffer, "echo")
	}
	if returnedBuffer != "echo" {
		t.Errorf("got returned buffer %q, want %q", returnedBuffer, "echo")
	}
	if err != nil {
		t.Errorf("got error %v, want nil", err)
	}
}
