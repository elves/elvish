package vals

import (
	"os"
)

// Pipe wraps a pair of [*os.File] that are the two ends of a pipe.
type Pipe struct{ R, W *os.File }
