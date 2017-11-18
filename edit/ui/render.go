package ui

// Renderer wraps the Render method.
type Renderer interface {
	// Render renders onto a Buffer.
	Render(b *Buffer)
}

// Render creates a new Buffer with the given width, and lets a Renderer render
// onto it.
func Render(r Renderer, width int) *Buffer {
	if r == nil {
		return nil
	}
	b := NewBuffer(width)
	r.Render(b)
	return b
}
