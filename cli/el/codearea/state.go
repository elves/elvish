package codearea

// State keeps the state of the widget. Its access must be synchronized through
// the mutex.
type State struct {
	Buffer  Buffer
	Pending Pending
}

// Buffer represents the state of the buffer.
type Buffer struct {
	// Content of the buffer.
	Content string
	// Position of the dot (more commonly known as the cursor), as a byte index
	// into Content.
	Dot int
}

// Pending represents pending code, such as during completion.
type Pending struct {
	// Beginning index of the text area that the pending code replaces, as a
	// byte index into RawState.Code.
	From int
	// End index of the text area that the pending code replaces, as a byte
	// index into RawState.Code.
	To int
	// The content of the pending code.
	Content string
}

// ApplyPending applies pending code to the code buffer, and resets pending code.
func (s *State) ApplyPending() {
	s.Buffer, _, _ = patchPending(s.Buffer, s.Pending)
	s.Pending = Pending{}
}

func (c *Buffer) InsertAtDot(text string) {
	*c = Buffer{
		Content: c.Content[:c.Dot] + text + c.Content[c.Dot:],
		Dot:     c.Dot + len(text),
	}
}
