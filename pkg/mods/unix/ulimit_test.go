//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

package unix

import (
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/mods/math"
)

func TestUlimit(t *testing.T) {
	const getHdrRegexp = `(?s)^resource .*\n======== .*\n`

	setup := func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddNs("unix", Ns))
		ev.ExtendGlobal(eval.BuildNs().AddNs("math", math.Ns))
	}
	TestWithSetup(t, setup,
		// Validate that invalid invocations produce the expected error.
		That(`unix:ulimit core 1 2`).Throws(errs.ArityMismatch{
			What: "arguments", ValidLow: 0, ValidHigh: 2, Actual: 3}),
		That(`unix:ulimit core 1/2`).Throws(ErrorWithType(vals.WrongType{})),
		That(`unix:ulimit core 0.5`).Throws(ErrorWithType(vals.WrongType{})),
		// This is an interesting case in as much as it's not obvious whether or not we
		// should allow converting a float64 that is a whole number to a uint64 if the value
		// has no fractional component and fits. This simply verifies the existing ScanToGo
		// conversion behavior at the time this test was written so we're alerted about a
		// potential backward incompatible change if this is ever allowed.
		That(`unix:ulimit core (float64 4096)`).Throws(ErrorWithType(vals.WrongType{})),
		// A 2^64 value, or larger, is invalid as a uint64. In practice a value larger than
		// 2^63-1 is probably invalid to the OS. This simply validates that a value too
		// large for a uint64 is rejected before we even attempt to use it.
		That(`unix:ulimit core (math:pow 2 64)`).Throws(ErrorWithType(vals.WrongType{})),
		// An unrecognized resource name is an error.
		That(`unix:ulimit invalid`).Throws(errs.BadValue{
			What:   "resource",
			Valid:  "in the output of ulimit:unix",
			Actual: "invalid"}),
		// Validate that asking for the current limits produces reasonable output.
		That(`unix:ulimit core | slurp`).Puts(
			MatchingRegexp{Pattern: getHdrRegexp +
				`core .* +(?:\d+|inf) +(?:\d+|inf)\n$`}),
		That(`unix:ulimit | slurp`).Puts(
			MatchingRegexp{Pattern: `(?s)^resource.*\n========.*\n` +
				`.*\ncore .* +(?:\d+|inf) +(?:\d+|inf)\n` +
				`.*\nnproc .* +(?:\d+|inf) +(?:\d+|inf)\n`}),

		// Validate that setting a limit works.
		//
		// NOTE: This is fragile. It assumes we can actually set the core file size limit to
		// 4KB and infinity which may not be true on every platform. Especially CI
		// environments. TBD is what to do about that if, and when, it is ever shown to be
		// an issue. Hopefully without having to stub the actual syscall(s).
		That(`unix:ulimit core 4096; unix:ulimit core | slurp`).Puts(
			MatchingRegexp{Pattern: getHdrRegexp +
				`core .* +4096 +(?:\d+|inf)\n$`}),
		That(`unix:ulimit core inf; unix:ulimit core | slurp`).Puts(
			MatchingRegexp{Pattern: getHdrRegexp +
				`core .* +inf +(?:\d+|inf)\n$`}),
	)
}
