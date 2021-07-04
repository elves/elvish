//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

package unix

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/sys/unix"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

type ulimitResource struct {
	human string // a human friendly description of the limit
	units string // the units for the limit
	id    int    // the setrlimit() syscall "resource" ID
}

//elvdoc:fn ulimit
//
// ```
// unix:ulimit [$resource [$value]]
// ```
//
// The `$resource` value is a string that identifies the resource limit. If no resource is specified
// all resource limits are written to the byte stream. Valid values depend on the platform and can
// be seen in the output of `unix:ulimit`.
//
// If no `$value` is provided the current and maximum allowed value for the resource limit is
// written to the byte stream. If a `$value` argument is provided the current (aka "soft") resource
// limit is modified. The maximum (aka "hard") limit cannot be modified. In other words, while
// traditional POSIX shells allow modifying the "soft" and "hard" limit this implementation only
// allows modifying the "soft" limit. The limit can be the special value `inf` if the maximum
// allowed is also `inf`.
//
// The unit of measurement for each resource limit is platform dependent but is usually the smallest
// meaningful unit. For example, the units are typically bytes for memory related resources such as
// the maximum size of a core dump file or the memory used by the process. Note that typical POSIX
// `ulimit` implementations scale some values by kbytes, blocks (whatever that means), etc. This
// implementation performs no scaling on input or output of the resource limits.
//
// ```elvish-transcript
// ~> unix:ulimit
// resource   description          units              current    maximum
// ========   ===========          =====              =======    =======
// as         virtual memory       bytes                  inf        inf
// core       core file size       bytes                    0        inf
// cpu        cpu time             seconds                inf        inf
// data       data segment size    bytes                  inf        inf
// fsize      file size            bytes                  inf        inf
// memlock    max locked memory    bytes                  inf        inf
// nofile     open files           count                  256        inf
// nproc      max user processes   count                 2088       2088
// rss        max memory size      bytes                  inf        inf
// stack      stack size           bytes              8388608   67104768
// ~> unix:ulimit core
// resource   description          units              current    maximum
// ========   ===========          =====              =======    =======
// core       core file size       bytes                    0        inf
// ~> unix:ulimit core (* 1024 (* 1024 500)) # set the limit to 500 MiB
// ~> unix:ulimit core
// resource   description          units              current    maximum
// ========   ===========          =====              =======    =======
// core       core file size       bytes            524288000        inf
// ```

func ulimit(fm *eval.Frame, vals ...interface{}) error {
	if len(vals) == 0 {
		return outputAllLimits(fm)
	}

	name, resource, err := validateResource(vals[0])
	if err != nil {
		return err
	}

	switch len(vals) {
	case 1:
		writeHdr(fm)
		return outputLimit(fm, name, resource)
	case 2:
		return setLimit(fm, name, resource, vals[1])
	default:
		return errs.ArityMismatch{What: "arguments",
			ValidLow: 0, ValidHigh: 2, Actual: len(vals)}
	}
}

func validateResource(name interface{}) (string, ulimitResource, error) {
	switch name := name.(type) {
	case string:
		resource, ok := ulimitResources[name]
		if !ok {
			return "", ulimitResource{}, errs.BadValue{What: "resource",
				Valid:  "in the output of ulimit:unix",
				Actual: name}
		}
		return name, resource, nil
	default:
		return "", ulimitResource{}, errs.BadValue{What: "resource",
			Valid:  "in the output of ulimit:unix",
			Actual: "kind " + vals.Kind(name)}
	}
}

// outputAllLimits does just what the name implies -- it's equivalent to POSIX "ulimit -a".
func outputAllLimits(fm *eval.Frame) error {
	writeHdr(fm)
	for _, name := range sortedResourceNames() {
		if err := outputLimit(fm, name, ulimitResources[name]); err != nil {
			return err
		}
	}
	return nil
}

// outputLimit outputs the current and maximum (aka "soft" and "hard") limits for the resource.
func outputLimit(fm *eval.Frame, name string, resource ulimitResource) error {
	var rlimit unix.Rlimit
	if err := unix.Getrlimit(resource.id, &rlimit); err != nil {
		return err // this should never happen
	}
	out := fm.ByteOutput()
	out.WriteString(fmtResource(name, resource, rlimit))
	return nil
}

// setLimit sets the current (aka "soft") limit for the named resource.
func setLimit(fm *eval.Frame, name string, resource ulimitResource, val interface{}) error {
	// Get the current resource limits since we don't want to change the maximum (aka "hard")
	// limit. We only want to change the current (aka "soft") limit.
	var rlimit unix.Rlimit
	if err := unix.Getrlimit(resource.id, &rlimit); err != nil {
		// This should never happen since the resource ID and rlimit struct should be valid.
		return err
	}

	// On some platforms (e.g., FreeBSD) the values in the unix.Rlimit struct are signed whereas
	// on most platforms they are unsigned. Hence the SetRlimitCur function to handle that
	// impedance mismatch rather than just doing vals.ScanToGo(v, &rlimit.Cur). In practice this
	// doesn't matter since the rlimit "infinity" value is usually 2^63-1 and thus can be
	// represented as a positive value by both a int64 and uint64 data type. Nonetheless, don't
	// assume that is true.
	newLimit, err := parseLimit(val)
	if err != nil {
		return err
	}
	SetRlimitCur(&rlimit, newLimit)
	return unix.Setrlimit(resource.id, &rlimit)
}

func parseLimit(val interface{}) (uint64, error) {
	var limit uint64
	switch val := val.(type) {
	case string:
		if val == "inf" {
			// We don't return math.MaxUint64 because on most platforms today the
			// "infinity" value is less than math.MaxUint64; specifically, 2^63-1.
			return unix.RLIM_INFINITY, nil
		}
		if err := vals.ScanToGo(val, &limit); err != nil {
			return 0, err
		}
		return limit, nil
	default:
		if err := vals.ScanToGo(val, &limit); err != nil {
			return 0, err
		}
		return limit, nil
	}
}

func sortedResourceNames() []string {
	resources := make([]string, 0)
	for key := range ulimitResources {
		resources = append(resources, key)
	}
	sort.Strings(resources)
	return resources
}

// The width of ten for the current and maximum values in the fmtResource() and writeHdr() functions
// accommodates 32 bit values. For now this is deemed "good enough" since most values, such as a
// core file size limit, could theoretically exceed 2^32 but is unlikely to do so unless it's the
// magic "inf" value. If the value is greater than 2^32 the fmtResource function will still render
// the value but without the expected vertical alignment. While visually unappealing it is still an
// accurate representation.

func fmtResource(name string, resource ulimitResource, rlimit unix.Rlimit) string {
	sb := new(strings.Builder)
	fmt.Fprintf(sb, "%-10s %-20s %-15s", name, resource.human, resource.units)
	if rlimit.Cur == unix.RLIM_INFINITY {
		fmt.Fprintf(sb, " %10s", "inf")
	} else {
		fmt.Fprintf(sb, " %10d", rlimit.Cur)
	}
	if rlimit.Max == unix.RLIM_INFINITY {
		fmt.Fprintf(sb, " %10s", "inf")
	} else {
		fmt.Fprintf(sb, " %10d", rlimit.Max)
	}
	sb.WriteByte('\n')
	return sb.String()
}

func writeHdr(fm *eval.Frame) {
	out := fm.ByteOutput()
	out.WriteString(fmt.Sprintf("%-10s %-20s %-15s %10s %10s\n",
		"resource", "description", "units", "current", "maximum"))
	out.WriteString(fmt.Sprintf("%-10s %-20s %-15s %10s %10s\n",
		"========", "===========", "=====", "=======", "======="))
}
