package cli

import "github.com/elves/elvish/cli/el/codearea"

// ReadCodeAsync is an asynchronous version of App.ReadCode. Instead of
// blocking, it returns immediately with two channels that will deliver the
// return values of ReadCode when ReadCode returns.
//
// This function is mainly useful in tests.
func ReadCodeAsync(a App) (<-chan string, <-chan error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, err := a.ReadCode()
		codeCh <- code
		errCh <- err
	}()
	return codeCh, errCh
}

// CodeBuffer returns the code buffer of the main code area widget of the app.
func CodeBuffer(a App) codearea.Buffer {
	return a.CodeArea().CopyState().Buffer
}

// SetCodeBuffer sets the code buffer of the main code area widget of the app.
func SetCodeBuffer(a App, buf codearea.Buffer) {
	a.CodeArea().MutateState(func(s *codearea.State) {
		s.Buffer = buf
	})
}
