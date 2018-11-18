package core

import "fmt"

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
