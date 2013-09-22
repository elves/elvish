// package async provides utilities for asynchronous IO.
package async

import (
	"errors"
)

var Timeout = errors.New("timed out")
