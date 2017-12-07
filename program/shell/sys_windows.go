package shell

import (
	"os"
)

func makeEditor(in, out *os.File, _ interface{}) *minEditor {
	return newMinEditor(in, out)
}

func handleSignal(_ os.Signal) {
}
