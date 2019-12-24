package codearea

// CodeAreaState keeps the mutable state of the CodeArea widget.
type CodeAreaState struct {
	Buffer      CodeBuffer
	Pending     PendingCode
	HideRPrompt bool
}

// CodeBuffer represents the buffer of the CodeArea widget.
type CodeBuffer struct {
	// Content of the buffer.
	Content string
	// Position of the dot (more commonly known as the cursor), as a byte index
	// into Content.
	Dot int
}

// PendingCode represents pending code, such as during completion.
type PendingCode struct {
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
func (s *CodeAreaState) ApplyPending() {
	s.Buffer, _, _ = patchPending(s.Buffer, s.Pending)
	s.Pending = PendingCode{}
}

func (c *CodeBuffer) InsertAtDot(text string) {
	*c = CodeBuffer{
		Content: c.Content[:c.Dot] + text + c.Content[c.Dot:],
		Dot:     c.Dot + len(text),
	}
}
