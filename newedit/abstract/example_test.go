package abstract

import "fmt"

func Example() {
	setuper := func() (func(), error) {
		fmt.Println("setup terminal")
		return func() { fmt.Println("restore terminal") }, nil
	}

	buffer := ""
	drawer := func(RedrawFlag) {
		fmt.Printf("buffer is %q\n", buffer)
	}
	input := func() (<-chan Event, func()) {
		events := make(chan Event, 5)
		for _, event := range "echo\n" {
			events <- event
		}
		return events, func() { fmt.Println("stop input") }
	}
	handler := func(event Event) (string, bool) {
		if event == '\n' {
			return buffer, true
		}
		buffer += string(event.(rune))
		return "", false
	}

	ed := NewEditor(input, handler)
	ed.SetupCb(setuper)
	ed.RedrawCb(drawer)
	ed.Read()
	// Output:
	// setup terminal
	// buffer is ""
	// buffer is "e"
	// buffer is "ec"
	// buffer is "ech"
	// buffer is "echo"
	// buffer is "echo"
	// stop input
	// restore terminal
}
