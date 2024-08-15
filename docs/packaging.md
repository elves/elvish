# Packaging Elvish

The main package of Elvish is `cmd/elvish`, and you can build it like any other
Go application.

## Enhancing version information

You can set some variables in the `src.elv.sh/pkg/buildinfo` package using
linker flags to enhance the Elvish's version information. See the
[package's API doc](https://pkg.go.dev/src.elv.sh@master/pkg/buildinfo) for
details.

They don't affect any other aspect of Elvish's behavior, so it's infeasible to
pass those linker flags, it's fine to leave them as is.

**Note**: The names and usage of these variables have changed several time in
Elvish's history. If your build script has `-ldflags '-X $symbol=$value'` where
`$symbol` is not documented in the linked API doc, those flags no longer do
anything and should be removed.

## Running tests

Some Elvish tests unfortunately rely on time thresholds. If you run tests as
part of the packaging process, you may want to set the
[`ELVISH_TEST_TIME_SCALE`](./testing.md#elvish_test_time_scale) environment
variable to a large value like 10.
