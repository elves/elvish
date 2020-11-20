// Package trace contains utilities useful for tracing the behavior of this project's Go code and
// programs written in Elvish.
package trace

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Class is used to classify a tracing message. This provides a means for controlling which messages
// are emitted at runtime.
type Class uint32

const (
	elvishModPrefix = "src.elv.sh/"
	maxNframes      = 25 // maximum number of stack frames to include in a backtrace
)

// This would be a constant but it needs to be modifiable by unit tests to verify the correct
// behavior of emitFrame().
var locationLenLimit int = 40

var mutex sync.Mutex              // serialize trace messages
var traceFH io.Writer             // current trace message stream
var defaultNframes int            // default number of stack frames to emit
var relativeTs bool               // whether a relative or absolute timestamp should be included
var localTs bool                  // whether a local or relative timestamp should be emitted
var now func() time.Time          // to facilitate unit test mocks
var baseTimes map[Class]time.Time // used to calculate a delta from the previous trace message
var enabledClasses Class          // bitmask of enabled message classes
var noBacktrace = errors.New("no backtrace available")

// InitState initialize the state of this module. This is meant to be used by unit tests that need
// to undo any changes by ParseTraceOptions, Printf, etc. Also called when this package is
// initialized at runtime. All of these initializations could be done statically but we would still
// need a function to reset the values between unit tests and would then have to keep this function
// in sync with the static initializations. Better to just do the initializations dynamically in
// both scenarios.
func InitState() {
	traceFH = os.Stderr
	defaultNframes = 5
	relativeTs = true
	localTs = true
	now = time.Now
	baseTimes = map[Class]time.Time{}
	enabledClasses = 0
}

// These constants must be kept in sync with the classToName map below.
const (
	// The Adhoc trace message class is meant to be used in ad-hoc trace.Printf calls that are
	// temporarily added by a developer. Do not leave such messages in merged changes.
	Adhoc = 1 << iota
	Cmd
	Daemon
	Eval
	Shell
	Store
	Terminal
)

// This must be kept in sync with the constants above.
//
// If the longest class name changes length the fmt.Fprintf calls below that use `[%-8s]` should be
// updated to keep things vertically aligned while minimizing trace message line length. Yes, the
// optimal width could be determined at runtime but for now I don't think that is justified since
// doing so makes tracing a trifle more expensive and it's not the end of the world if the padding
// width is not optimal.
var classToName = map[Class]string{
	Adhoc:    "adhoc",
	Cmd:      "cmd",
	Daemon:   "daemon",
	Eval:     "eval",
	Shell:    "shell",
	Store:    "store",
	Terminal: "terminal",
}

func (n Class) String() string {
	if name, ok := classToName[n]; ok {
		return name
	}
	return "Class(0b" + strconv.FormatUint(uint64(n), 2) + ")"
}

func nameToClass(maybeName string) (Class, error) {
	for class, name := range classToName {
		if maybeName == name {
			return class, nil
		}
	}
	return 0, fmt.Errorf("unknown trace class: %s", maybeName)
}

// IsEnabled returns true if the provided class is enabled for tracing.
func IsEnabled(class Class) bool {
	return class&enabledClasses == class
}

// ParseTraceOptions parses a `-trace` option value and configures the tracing subsystem
// accordingly. The trace string contains tokens separated by spaces or commas. The recognized
// tokens are:
//
//   file=     The string after the equal-sign is a pathname to open for writing.
//             It's often useful to use the /dev/tty device of another terminal.
//             The default is the stderr file descriptor.
//   nframes=  The number of backtrace frames to include in messages that don't specifify an
//             explicit number of frames. The default is 5. A negative value, or a value greater
//             than the maximum allowed, will be changed to the maximum allowed (see maxNframes).
//   local     Include a timestamp in the local timezone in YYYY-MM-DD HH:MM:SS.mmmmmm format.
//             The default is the number of seconds (and microseconds) since the
//             previous trace message was written.
//   utc       Include a timestamp in the UTC timezone in YYYY-MM-DD HH:MM:SS.mmmmmm format.
//   all       Enable all trace message classes.
//
// You can also specify specific trace message classes to be emitted to
// minimize unwanted noise in the trace output.
//
//   adhoc
//   cmd
//   daemon
//   eval
//   shell
//   store
//   terminal
//
func ParseTraceOptions(options string) error {
	var reterr error // we'll return nil or the last error seen parsing the trace options
	fields := strings.FieldsFunc(options, func(r rune) bool { return r == ' ' || r == ',' })
	for _, arg := range fields {
		if strings.HasPrefix(arg, "file=") {
			file := arg[len("file="):]
			if fh, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				traceFH = fh
			} else {
				reterr = fmt.Errorf("cannot open trace file for writing: %s", file)
			}
		} else if strings.HasPrefix(arg, "nframes=") {
			nframes, err := strconv.ParseInt(arg[len("nframes="):], 0, 0)
			if err == nil {
				if nframes < 0 || nframes > maxNframes {
					defaultNframes = maxNframes
				} else {
					defaultNframes = int(nframes)
				}
			} else {
				reterr = err
			}
		} else if arg == "local" {
			relativeTs = false // disable relative, and enable absolute, timestamps
			localTs = true     // printed in the local timezone
		} else if arg == "utc" {
			relativeTs = false // disable relative, and enable absolute, timestamps
			localTs = false    // printed in the UTC timezone
		} else if arg == "all" {
			for class := range classToName {
				enabledClasses |= class
			}
		} else if class, err := nameToClass(arg); err == nil {
			enabledClasses |= class
		} else {
			reterr = fmt.Errorf("unknown trace option: %s", arg)
		}
	}
	return reterr
}

func getTimestamp(class Class) string {
	if !relativeTs {
		// Return an absolute timestamp with microsecond resolution.
		if localTs {
			return now().Local().Format("2006-01-02 15:04:05.000000")
		}
		return now().UTC().Format("2006-01-02 15:04:05.000000")
	}

	// Return a relative timestamp, with microsecond resolution, as a string since the previous
	// message for the class was emitted. This is relative to the message class since we typically
	// want to know when the previous message of the message class was emitted. For example, when
	// the previous message involving signals was emitted.
	t := now()
	baseTime, _ := baseTimes[class]
	if baseTime.IsZero() { // this is our first trace message for the message class
		baseTime = t
	}
	delta := t.Sub(baseTime)
	baseTimes[class] = t
	return fmt.Sprintf("%10.6f", delta.Seconds())
}

// Printf writes a tracing message to the current tracing stream with an optional backtrace. Whether
// the tracing message is emitted is controlled by which trace message classes are enabled.
//
// class:   The numeric trace class the message belongs to.
// nframes: 0 to include no backtrace in the message
//          1 to include only our caller's stack frame
//          >1 to include the N most recent stack frames
//          <0 to report the default number of stack frames (see `defaultNframes`)
// format:  A fmt.Printf style formatting string. Can contain embedded newlines to
//          produce multiple lines of output. A trailing newline will be ignored.
// args:    Optional values for the `format` argument.
func Printf(class Class, nframes int, format string, args ...interface{}) {
	if IsEnabled(class) {
		DoPrintf(class, 1, nframes, format, args...)
	}
}

// DoPrintf is responsible for the actual output of an enabled tracing message. We deliberately
// break the implementation in two pieces so that `trace.Printf` can be inlined and the inlined code
// is as small as possible. Thus, avoiding a function call for the common case where tracing is not
// enabled.
func DoPrintf(class Class, skip, nframes int, format string, args ...interface{}) {
	mutex.Lock()
	defer mutex.Unlock()

	ts := getTimestamp(class)
	msg := fmt.Sprintf(format, args...)
	lines := strings.Split(msg, "\n")
	if len(lines) > 1 && lines[len(lines)-1] == "" { // message ended with a newline
		lines = lines[:len(lines)-1] // drop the empty line created by the trailing newline
	}

	for i, msgLine := range lines {
		fmt.Fprintf(traceFH, "%s [%-8s] %s\n", ts, class, msgLine)
		if i == 0 {
			ts = fmt.Sprintf("%*s", len(ts), "...")
		}
	}

	if frames, err := getBacktrace(skip+1, nframes); err == nil {
		frame, more := frames.Next()
		emitFrame(ts, class, 1, frame)
		for fno := 2; more; fno++ {
			frame, more = frames.Next()
			emitFrame(ts, class, fno, frame)
		}
	}
}

func getBacktrace(skip, nframes int) (*runtime.Frames, error) {
	if nframes == 0 {
		return nil, noBacktrace
	}
	if nframes < 0 {
		nframes = defaultNframes
	}

	pc := make([]uintptr, nframes)
	// The "+ 2" is because we want to skip our frame and runtime.Callers. The call chain might look
	// like this:
	//   target func -> trace.Printf -> trace.DoPrintf -> trace.getBacktrace -> runtime.Callers
	// We want to ignore the last four frames as they aren't interesting.
	n := runtime.Callers(skip+2, pc)
	if n == 0 { // should never be true but don't panic if it happens
		return nil, noBacktrace
	}
	if n != nframes {
		pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
	}
	return runtime.CallersFrames(pc), nil
}

func emitFrame(ts string, class Class, fno int, frame runtime.Frame) {
	line := frame.Line
	file := frame.File
	if n := strings.Index(file, "/elvish/"); n != -1 {
		file = file[n+len("/elvish/"):]
	}

	//	if strings.HasPrefix(file, elvishModPrefix) {
	//		file = file[len(elvishModPrefix):]
	//	}
	location := fmt.Sprintf("%s:%d", file, line)

	funcname := frame.Func.Name()
	if strings.HasPrefix(funcname, elvishModPrefix) {
		funcname = funcname[len(elvishModPrefix):]
	}
	funcname += "()"

	// Keep function names vertically aligned for readability.
	if len(location) > locationLenLimit {
		fmt.Fprintf(traceFH, "%s [%-8s] %2d# %s\n", ts, class, fno, location)
		location = ""
	}
	fmt.Fprintf(traceFH, "%s [%-8s] %2d# %-*s  %s\n", ts, class, fno, locationLenLimit, location,
		funcname)
}

func init() {
	InitState()
}
