package util

// PPrinter wraps the PPrint function.
type PPrinter interface {
	// PPrint takes an indentation string and pretty-prints.
	PPrint(indent string) string
}
