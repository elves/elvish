package errutil

import "fmt"

type PosError struct {
	Begin int
	End   int
	Err   error
}

func (pe *PosError) Error() string {
	return fmt.Sprintf("%d-%d: %s", pe.Begin, pe.End, pe.Err.Error())
}
