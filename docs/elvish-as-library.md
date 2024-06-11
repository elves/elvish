# Using Elvish as a library

Elvish's implementation is structured as a collection of Go packages with
well-documented internal APIs, so it's possible to use the parts you're
interested in as a Go library.

-   Most likely, you'll want to use Elvish's interpreter. The examples for the
    [`Evaler.Eval` method](https://pkg.go.dev/src.elv.sh@master/pkg/eval#Evaler.Eval)
    should give you a good starting point.

-   For a general overview of how Elvish's code is structured, read the
    [architecture overview](https://pkg.go.dev/src.elv.sh@master/docs/architecture).

However, beware that Elvish promises no backward compatibility in its Go API.
The internal API surface is large, and will change from time to time as Elvish's
implementation gets refactored.

For now, this is consistent with Go's semantic versioning rules as Elvish is
pre-1.0. When Elvish 1.0 is eventually released, all the internal libraries will
likely be moved into an `internal` directory, with a small part of the API
exposed via facades in the `pkg` directory.
