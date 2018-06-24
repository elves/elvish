package loop

import "fmt"

func Example() {
	buffer := ""
	firstDrawerCall := true
	drawer := func(flag RedrawFlag) {
		// Because the consumption of events is batched, calls to the drawer is
		// nondeterministic except for the first and final calls.
		switch {
		case firstDrawerCall:
			fmt.Printf("initial buffer is %q\n", buffer)
			firstDrawerCall = false
		case flag&FinalRedraw != 0:
			fmt.Printf("final buffer is %q\n", buffer)
		}
	}
	handler := func(event Event) (string, bool) {
		if event == '\n' {
			return buffer, true
		}
		buffer += string(event.(rune))
		return "", false
	}

	ed := New()
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
