package diag

// Shower wraps the Show function.
type Shower interface {
	// Show takes an indentation string and shows.
	Show(indent string) string
}
