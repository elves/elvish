# Testing changes

Write comprehensive unit tests for your code, and make sure that existing tests
are passing. Run tests with `make test`.

Respect established patterns of how unit tests are written. Some packages
unfortunately have competing patterns, which usually reflects a still-evolving
idea of how to best test the code. Worse, parts of the codebase are poorly
tested, or even untestable. In either case, discuss with the project lead on the
best way forward.

### Transcript tests

Most tests against Elvish modules are written in `.elvts` files, which mimic
transcripts of Elvish REPL sessions. See
https://pkg.go.dev/src.elv.sh@master/pkg/transcript for the format of transcript
files, and https://pkg.go.dev/src.elv.sh@master/pkg/eval/evaltest for details
specific to using them as tests.

If you use VS Code, the official Elvish extension allows you to simply press
<kbd>Alt-Enter</kbd> to update the output of transcripts (specifically, the
output for the code block the cursor is in). This means that you can author
transcript tests entirely within the editor, instead of manually writing out the
expected output or copy-pasting outputs from an actual REPL.

**Note**: The functionality of the VS Code plugin is based on a very simple
protocol and can be easily implemented for other editors. The protocol is
documented in the godoc for the `evaltest` package (see link below), and you can
also take a look `vscode/src/extension.ts` for the client implementation in the
VS Code extension.

### ELVISH_TEST_TIME_SCALE

Some unit tests depend on time thresholds. The default values of these time
thresholds are suitable for a reasonably powerful laptop, but on
resource-constraint environments (virtual machines, embedded systems) they might
not be enough.

Set the `ELVISH_TEST_TIME_SCALE` environment variable to a number greater than 1
to scale up the time thresholds used in tests. The CI environments use
`ELVISH_TEST_TIME_SCALE = 10`.

### Mocking dependencies

Whenever possible, test the real thing.

However, there are situations where it's infeasible to test the real thing, like
syscall errors that can't be reliably triggered, or tests that rely on exact
timing. In those cases, introduce a variable that stores the actual dependency
(manual dependency injection):

```go
// f.go
package pkg

import "os"

var osSleep = os.Sleep

func F() {
    // Use osSleep instead of os.Sleep
}
```

And then use `testutil.Set` to override it for the duration of a test:

```go
// f_test.go
package pkg

import "testing"

func TestF(t *testing.T) {
    testutil.Set(&osSleep, func(d Duration) {
        // Fake implementation
    })
    // Now test F
}
```

If the test is in an external test package, the dependency variable will have to
be exported. Instead of exporting it directly in the implementation file, export
a pointer to it in a internal test file:

```go
// testexport_test.go
package pkg // Note: internal

var OSSleep = &os.Sleep

// f_test.go
package pkg_test // Note: external

import (
    "pkg"
    "testing"
)

func TestF(t *testing.T) {
    // Note: No more & since pkg.OSSleep is already a pointer
    testutil.Set(pkg.OSSleep, func(d Duration) {
        // Fake implementation
    })
    // Now test F
}
```
