package eval

import (
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/trace"
)

// Ideally, the code in this file would be part of the `trace` package. It can't because that would
// create an import cycle. Specifically, we want the `eval` package to be able to call functions in
// the `trace` package (e.g., `trace.Printf`). Yet the `Frame` type is defined by the `eval` package
// so we can't import that package into the `trace` package without creating an import cycle.

// fmPrintln is a specialization of trace.Printf. The key difference is the inclusion of an
// eval.Frame pointer in the arguments so we can include Elvish source context in the message. We
// also hope that by keeping this function extremely simple this is inlined to avoid an unwanted
// function call when tracing a message class that is not enabled.
func fmPrintln(fm *Frame, class trace.Class, r diag.Ranger, msg string, args ...interface{}) {
	if trace.IsEnabled(class) {
		doFmPrintln(class, fm, r, msg, args)
	}
}

// As with trace.Printf we deliberately break this into two functions to maximize the probability
// fmPrintln will be inlined; thus avoiding a function call in the common case where tracing is not
// enabled.
func doFmPrintln(class trace.Class, fm *Frame, r diag.Ranger, msg string, args []interface{}) {
	// Depending on how this function was invoked the current context may not have been added to the
	// frame's traceback list. So ensure the current context reflects the code being executed.
	tb := fm.addTraceback(r)
	location, source := tb.Head.Description()
	// We don't want an arbitrary number of source lines in the trace output. Output just the first
	// three lines. Most ops span a single line. The few that don't tend to be things like `for`
	// blocks where we don't care about the body when printing a trace message for an iteration.
	// This is a very crude attempt to minimize the trace output to the essential information.
	lines := strings.Split(source, "\n")
	if len(lines) > 3 {
		source = strings.Join(lines[:3], "\n")
	}

	if msg == "" && len(args) > 0 { // caller really should pass a non-empty string in this case
		msg = "ARGS"
	}
	if msg != "" {
		if len(args) == 0 {
			trace.DoPrintf(class, 2, 0, "@ %s\n%s\n%s", location, source, msg)
		} else {
			// We don't want the caller of fmPrintln to incur the cost of calling vals.Repr if
			// tracing is not enabled --- so we do it here.
			trace.DoPrintf(class, 2, 0, "@ %s\n%s\n%s: %s", location, source, msg,
				vals.Repr(args, vals.NoPretty))
		}
	} else {
		trace.DoPrintf(class, 2, 0, "@ %s\n%s", location, source)
	}
}
