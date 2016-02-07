package errutil

import "fmt"

// PosError is an error associated with a position range.
type PosError struct {
	Begin int
	End   int
	Err   error
}

func (pe *PosError) Error() string {
	return fmt.Sprintf("%d-%d: %s", pe.Begin, pe.End, pe.Err.Error())
}
