package edit

import "github.com/elves/elvish/eval"

// getIsExternal finds a set of all external commands and puts it on the result
// channel.
func getIsExternal(ev *eval.Evaler, result chan<- map[string]bool) {
	names := make(chan string, 32)
	go func() {
		ev.AllExecutables(names)
		close(names)
	}()
	isExternal := make(map[string]bool)
	for name := range names {
		isExternal[name] = true
	}
	result <- isExternal
}
