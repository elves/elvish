package abstract

import "fmt"

func Example() {
	buffer := ""
	drawer := func(RedrawFlag) {
		fmt.Printf("buffer is %q\n", buffer)
	}
	handler := func(event Event) (string, bool) {
		if event == '\n' {
			return buffer, true
		}
		buffer += string(event.(rune))
		return "", false
	}

	ed := NewEditor(handler)
	go func() {
		for _, event := range "echo\n" {
			ed.Input(event)
		}
	}()
	ed.RedrawCb(drawer)
	ed.Read()
	// Output:
	// buffer is ""
	// buffer is "e"
	// buffer is "ec"
	// buffer is "ech"
	// buffer is "echo"
	// buffer is "echo"
}
