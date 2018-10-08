package util

// Pprinter wraps the Pprint function.
type Pprinter interface {
	// Pprint takes an indentation string and pretty-prints.
	Pprint(indent string) string
}
