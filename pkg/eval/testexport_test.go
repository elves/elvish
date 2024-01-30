package eval

// Pointers to variables that can be mutated for testing.
var (
	GetHome       = &getHome
	Getwd         = &getwd
	OSExit        = &osExit
	TimeAfter     = &timeAfter
	TimeNow       = &timeNow
	NextEvalCount = &nextEvalCount

	ExceptionCauseStartMarker = &exceptionCauseStartMarker
	ExceptionCauseEndMarker   = &exceptionCauseEndMarker
)
