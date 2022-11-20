//go:build !windows && !plan9 && !js

package unix

import (
	"fmt"
	"math"
	"strconv"
	"sync"

	"golang.org/x/sys/unix"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

const (
	validUmaskMsg = "integer in the range [0..0o777]"
)

//elvdoc:var umask
//
// The file mode creation mask. Its value is a string in Elvish octal
// representation; e.g. 0o027. This makes it possible to use it in any context
// that expects a `$number`.
//
// When assigning a new value a string is implicitly treated as an
// octal number. If that fails the usual rules for interpreting
// [numbers](./language.html#number) are used. The following are equivalent:
// `set unix:umask = 027` and `set unix:umask = 0o27`. You can also assign to it
// a `float64` data type that has no fractional component. The assigned value
// must be within the range [0 ... 0o777], otherwise the assignment will throw
// an exception.
//
// You can do a temporary assignment to affect a single command; e.g. `umask=077
// touch a_file`. After the command completes the old umask will be restored.
// **Warning**: Since the umask applies to the entire process, not individual
// threads, changing it temporarily in this manner is dangerous if you are doing
// anything in parallel, such as via the [`peach`](builtin.html#peach) command.

// UmaskVariable is a variable whose value always reflects the current file
// creation permission mask. Setting it changes the current file creation
// permission mask for the process (not an individual thread).
type UmaskVariable struct{}

var _ vars.Var = UmaskVariable{}

// There is no way to query the current umask without changing it, so we store
// the umask value in a variable, and initialize it during startup. It needs to
// be mutex-guarded since it could be read or written concurrently.
//
// This assumes no other part of the Elvish code base involved in the
// interpreter ever calls unix.Umask, which is guaranteed by the
// check-content.sh script.
var (
	umaskVal   int
	umaskMutex sync.RWMutex
)

func init() {
	// Init functions are run concurrently, so it's normally impossible to
	// observe the temporary value.
	//
	// Even if there is some pathological init logic (e.g. goroutine from init
	// functions), the failure pattern is relative safe because we are setting
	// the temporary umask to the most restrictive value possible.
	umask := unix.Umask(0o777)
	unix.Umask(umask)
	umaskVal = umask
}

// Get returns the current file creation umask as a string.
func (UmaskVariable) Get() any {
	umaskMutex.RLock()
	defer umaskMutex.RUnlock()
	return fmt.Sprintf("0o%03o", umaskVal)
}

// Set changes the current file creation umask. It can be called with a string
// or a number. Strings are treated as octal numbers by default, unless they
// have an explicit base prefix like 0x or 0b.
func (UmaskVariable) Set(v any) error {
	umask, err := parseUmask(v)
	if err != nil {
		return err
	}

	umaskMutex.Lock()
	defer umaskMutex.Unlock()
	unix.Umask(umask)
	umaskVal = umask
	return nil
}

func parseUmask(v any) (int, error) {
	var umask int

	switch v := v.(type) {
	case string:
		i, err := strconv.ParseInt(v, 8, 0)
		if err != nil {
			i, err = strconv.ParseInt(v, 0, 0)
			if err != nil {
				return -1, errs.BadValue{
					What: "umask", Valid: validUmaskMsg, Actual: vals.ToString(v)}
			}
		}
		umask = int(i)
	case int:
		umask = v
	case float64:
		intPart, fracPart := math.Modf(v)
		if fracPart != 0 {
			return -1, errs.BadValue{
				What: "umask", Valid: validUmaskMsg, Actual: vals.ToString(v)}
		}
		umask = int(intPart)
	default:
		// We don't bother supporting big.Int or bit.Rat because no valid umask
		// value would be represented by those types.
		return -1, errs.BadValue{
			What: "umask", Valid: validUmaskMsg, Actual: vals.Kind(v)}
	}

	if umask < 0 || umask > 0o777 {
		return -1, errs.OutOfRange{
			What: "umask", ValidLow: "0", ValidHigh: "0o777",
			Actual: fmt.Sprintf("%O", umask)}
	}
	return umask, nil
}
