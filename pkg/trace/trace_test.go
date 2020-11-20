package trace

import (
	"bytes"
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"src.elv.sh/pkg/testutil"
)

// Change some of the trace defaults to something friendlier for unit tests.
func initTraceState() {
	defaultNframes = 3
	relativeTs = true
	localTs = false
	one_sec, _ := time.ParseDuration("1s") // this can't fail so skip err check
	t := time.Time{}
	now = func() time.Time { // this returns a time in the UTC location
		t = t.Add(one_sec)
		return t
	}
}

func TestParseTraceOptions(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()
	defer InitState()
	initTraceState()

	// An empty trace option string should do nothing and not return an error.
	err := ParseTraceOptions("")
	if err != nil {
		t.Errorf(`ParseTraceOptions("") returned unexpected error: %v`, err)
	}
	if traceFH != os.Stderr {
		t.Errorf(`ParseTraceOptions("") changed traceFH`)
	}
	if defaultNframes != 3 {
		t.Errorf(`ParseTraceOptions("") changed defaultNframes: exp %d, got %d`, 5, defaultNframes)
	}

	// One or more invalid trace options returns an error for the last invalid
	// option but otherwise processes valid options. This also verifies the
	// various ways to separate trace tokens are handled correctly.
	opts := "local,unk1,nframes=3,unk2  adhoc,cmd, file=trace.log eval"
	err = ParseTraceOptions(opts)
	exp := errors.New("unknown trace option: unk2")
	if err.Error() != exp.Error() {
		t.Errorf("ParseTraceOptions(%q): exp %q, got %q", opts, exp, err)
	}
	if relativeTs != false {
		t.Errorf("ParseTraceOptions(%q) should have disabled relativeTs", opts)
	}
	if localTs != true {
		t.Errorf("ParseTraceOptions(%q) should have enabled localTs", opts)
	}
	if traceFH == os.Stderr {
		t.Errorf("ParseTraceOptions(%q) should have changed traceFH", opts)
	}
	if defaultNframes != 3 {
		t.Errorf("ParseTraceOptions(%q) changed defaultNframes: exp %d, got %d",
			opts, 3, defaultNframes)
	}
	if enabledClasses != Adhoc|Cmd|Eval {
		t.Errorf("ParseTraceOptions(%q) wrong enabledClasses", opts)
	}

	// An invalid trace file should return an error.
	opts = "file=/argle/bargle"
	err = ParseTraceOptions(opts)
	exp = errors.New("cannot open trace file for writing: /argle/bargle")
	if err == nil || err.Error() != exp.Error() {
		t.Errorf("ParseTraceOptions(%q): exp %q, got %q", opts, exp, err)
	}

	// An invalid trace nframes integer should return an error whereas a negative value should use
	// the max allowed.
	opts = "nframes=invalid-int"
	err = ParseTraceOptions(opts)
	if err == nil {
		t.Errorf("ParseTraceOptions(%q) should have returned an error", opts)
	}
	opts = "nframes=-1"
	err = ParseTraceOptions(opts)
	if err != nil {
		t.Errorf("ParseTraceOptions(%q) should not have returned an error", opts)
	}
	if defaultNframes != maxNframes {
		t.Errorf("ParseTraceOptions(%q) wrong: exp %d, act %d", opts, maxNframes, defaultNframes)
	}
	opts = "nframes=0"
	err = ParseTraceOptions(opts)
	if err != nil {
		t.Errorf("ParseTraceOptions(%q) should not have returned an error", opts)
	}
	if defaultNframes != 0 {
		t.Errorf("ParseTraceOptions(%q) wrong: exp %d, act %d", opts, 0, defaultNframes)
	}

	// Verify the "all" trace option is handled correctly. We include a couple
	// of other valid options that should not affect the result of using "all"
	// just because we're paranoid.
	opts = "utc all eval"
	err = ParseTraceOptions(opts)
	if err != nil {
		t.Errorf("ParseTraceOptions(%q) should not have returned an error", opts)
	} else {
		var exp Class
		for class := range classToName {
			exp |= class
		}
		if enabledClasses != exp {
			t.Errorf("ParseTraceOptions(%q) wrong enabledClasses: exp %b, act %b",
				opts, exp, enabledClasses)
		}
	}
	if localTs != false {
		t.Errorf("ParseTraceOptions(%q) should have disabled localTs", opts)
	}
}

func TestTraceGetTimestamp(t *testing.T) {
	var ts string
	defer InitState()
	initTraceState()

	relativeTs = true
	ts = getTimestamp(Cmd)
	if ts != "  0.000000" {
		t.Errorf("getTimestamp() unexpected result: %s", ts)
	}
	ts = getTimestamp(Cmd)
	if ts != "  1.000000" {
		t.Errorf("getTimestamp() unexpected result: %s", ts)
	}

	initTraceState()
	relativeTs = false
	ts = getTimestamp(Cmd)
	if ts != "0001-01-01 00:00:01.000000" {
		t.Errorf("getTimestamp() unexpected result: %s", ts)
	}

	// If the default timezone is not UTC verify that getTimestamp() does not
	// return a timestamp formatted in the UTC timezone. Yes, this is a hack,
	// but it is hard to justify anything more complicated. We hope most
	// Elvish test environments use a local timezone that is not UTC.
	_, offset := time.Now().Zone()
	if offset != 0 {
		localTs = true
		relativeTs = false
		ts = getTimestamp(Cmd)
		if ts == "0001-01-01 00:00:00.000000" {
			t.Errorf("getTimestamp() unexpected result: %s", ts)
		}
	}
}

func TestTraceClassStringer(t *testing.T) {
	var class Class

	class = 0b11 // classes are bit masks with a single bit set so this is invalid
	if class.String() != "Class(0b11)" {
		t.Errorf("Class.String() unexpected result: %s", class.String())
	}
	class = Adhoc
	if class.String() != "adhoc" {
		t.Errorf("Class.String() unexpected result: %s", class.String())
	}
	class = Shell
	if class.String() != "shell" {
		t.Errorf("Class.String() unexpected result: %s", class.String())
	}
}

func TestTracePrintf(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()
	defer InitState()
	initTraceState()

	var exp string
	var buf bytes.Buffer
	traceFH = &buf

	// Printing to a trace class that is not enabled should produce no output.
	buf.Reset()
	Printf(Cmd, 0, "should not be written")
	if buf.Len() != 0 {
		t.Errorf("Printf(Cmd,...) unexpected output: %s", buf.String())
	}

	// Printing to a trace class that is enabled should produce output.
	buf.Reset()
	ParseTraceOptions("adhoc,cmd")
	Printf(Cmd, 0, "should be written")
	if s := buf.String(); !strings.HasSuffix(s, " should be written\n") {
		t.Errorf("Printf(Cmd,...) unexpected output: %s", s)
	}

	// Printing to the Adhoc class is always enabled.
	buf.Reset()
	Printf(Adhoc, 0, "adhoc msg")
	if s := buf.String(); !strings.HasSuffix(s, " adhoc msg\n") {
		t.Errorf("Printf(Adhoc,...) unexpected output: %s", s)
	}

	// A trace message with a trailing newline should not produce an empty
	// line since trailing newlines in trace messages are not required and
	// ignored.
	buf.Reset()
	Printf(Adhoc, 0, "adhoc msg: %v\n", true)
	if s := buf.String(); !strings.HasSuffix(s, " adhoc msg: true\n") {
		t.Errorf("Printf(Adhoc,...) unexpected output:\n%s", s)
	}

	// A multiline trace message is correctly wrapped.
	buf.Reset()
	Printf(Adhoc, 0, "line 1\nline 2\n")
	exp = "  1.000000 [adhoc   ] line 1\n       ... [adhoc   ] line 2\n"
	if s := buf.String(); s != exp {
		t.Errorf("Printf(Adhoc,...) unexpected output:\nexp: %q\nact: %q", exp, s)
	}

	// A trace message with one backtrace frame produces the expected output.
	//
	// This is potentially fragile since it relies on inlining the `trace.Printf` function. On the
	// other hand, since we expect it to be inlined seeing this test fail is cause for concern.
	buf.Reset()
	locationLenLimit = 0
	Printf(Adhoc, 1, "one frame")
	matched, err := regexp.Match(`^  1.000000 \[adhoc   \] one frame\n`+
		`.* 1# .*pkg/trace/trace_test\.go:.*\n`+
		`.* 1#   .*\(\)\n$`,
		buf.Bytes())
	if !matched || err != nil {
		t.Errorf("Printf(Adhoc,...) unexpected output:\n%s", buf.String())
	}

	// Force the location and function names to be on different lines for every frame. Also verify a
	// corner case involving the translation of a negative frame count to a default number of frames
	// that is larger than the actual number of frames available.
	//
	// This is potentially fragile since it relies on how inlining the `trace.Printf` function
	// affects the output of the runtime.Callers() function.
	buf.Reset()
	locationLenLimit = 0
	defaultNframes = 999
	Frame2()
	matched, err = regexp.Match(`^.* default frames\n`+
		`.* 1# .*pkg/trace/trace_test\.go:.*\n`+
		`.* 1#   .*\(\)\n`+
		`.* 2# .*\n`+
		`.* 2#   .*\(\)\n`+
		`.* 3# .*pkg/trace/trace_test\.go:.*\n`+
		`.* 3#   pkg/trace.TestTracePrintf\(\)\n`,
		buf.Bytes())
	if !matched || err != nil {
		t.Errorf("Printf(Adhoc,...) unexpected output:\n%s", buf.String())
	}
}

func Frame2() {
	Frame3()
}

func Frame3() {
	Printf(Adhoc, -1, "default frames")
}
