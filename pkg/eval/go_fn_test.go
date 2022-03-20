package eval_test

import (
	"errors"
	"math/big"
	"reflect"
	"testing"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
)

type someOptions struct {
	Foo string
	Bar string
}

func (o *someOptions) SetDefaultOptions() { o.Bar = "default" }

//lint:ignore ST1012 test code
var anError = errors.New("an error")

type namedSlice []string

func TestGoFn_RawOptions(t *testing.T) {
	Test(t,
		That("f").DoesNothing().
			WithSetup(f(func() {})),

		// RawOptions
		That("f &foo=bar").DoesNothing().
			WithSetup(f(func(opts RawOptions) {
				if opts["foo"] != "bar" {
					t.Errorf("RawOptions parameter doesn't get options")
				}
			})),
		// Options when the function does not accept options.
		That("f &foo=bar").Throws(ErrNoOptAccepted).
			WithSetup(f(func() {
				t.Errorf("Function called when there are extra options")
			})),
		// Parsed options
		That("f &foo=bar").DoesNothing().
			WithSetup(f(func(opts someOptions) {
				if opts.Foo != "bar" {
					t.Errorf("ScanOptions parameter doesn't get options")
				}
				if opts.Bar != "default" {
					t.Errorf("ScanOptions parameter doesn't use default value")
				}
			})),
		// Invalid option; regression test for #958.
		That("f &bad=bar").Throws(UnknownOption{"bad"}).
			WithSetup(f(func(opts someOptions) {
				t.Errorf("function called when there are invalid options")
			})),
		// Invalid option type; regression test for #958.
		That("f &foo=[]").Throws(ErrorWithType(vals.WrongType{})).
			WithSetup(f(func(opts someOptions) {
				t.Errorf("function called when there are invalid options")
			})),

		// Argument
		That("f lorem ipsum").DoesNothing().
			WithSetup(f(func(x, y string) {
				if x != "lorem" {
					t.Errorf("Argument x not passed")
				}
				if y != "ipsum" {
					t.Errorf("Argument y not passed")
				}
			})),
		// Variadic arguments
		That("f lorem ipsum").DoesNothing().
			WithSetup(f(func(args ...string) {
				wantArgs := []string{"lorem", "ipsum"}
				if !reflect.DeepEqual(args, wantArgs) {
					t.Errorf("got args %v, want %v", args, wantArgs)
				}
			})),
		// Argument conversion
		That("f 314 1.25").DoesNothing().
			WithSetup(f(func(i int, f float64) {
				if i != 314 {
					t.Errorf("Integer argument i not passed")
				}
				if f != 1.25 {
					t.Errorf("Float argument f not passed")
				}
			})),
		// Inputs
		That("f [foo bar]").DoesNothing().
			WithSetup(f(testInputs(t, "foo", "bar"))),
		That("f [foo bar]").DoesNothing().
			WithSetup(f(testInputs(t, "foo", "bar"))),
		// Too many arguments
		That("f x").
			Throws(errs.ArityMismatch{What: "arguments",
				ValidLow: 0, ValidHigh: 0, Actual: 1}).
			WithSetup(f(func() {
				t.Errorf("Function called when there are too many arguments")
			})),
		// Too few arguments
		That("f").
			Throws(errs.ArityMismatch{What: "arguments",
				ValidLow: 1, ValidHigh: 1, Actual: 0}).
			WithSetup(f(func(x string) {
				t.Errorf("Function called when there are too few arguments")
			})),
		That("f").
			Throws(errs.ArityMismatch{What: "arguments",
				ValidLow: 1, ValidHigh: -1, Actual: 0}).
			WithSetup(f(func(x string, y ...string) {
				t.Errorf("Function called when there are too few arguments")
			})),
		// Wrong argument type
		That("f (num 1)").Throws(ErrorWithType(WrongArgType{})).
			WithSetup(f(func(x string) {
				t.Errorf("Function called when arguments have wrong type")
			})),
		That("f str").Throws(ErrorWithType(WrongArgType{})).
			WithSetup(f(func(x int) {
				t.Errorf("Function called when arguments have wrong type")
			})),

		// Return value
		That("f").Puts("foo").
			WithSetup(f(func() string { return "foo" })),
		// Return value conversion
		That("f").Puts(314).
			WithSetup(f(func() *big.Int { return big.NewInt(314) })),
		// Slice and array return value
		That("f").Puts("foo", "bar").
			WithSetup(f(func() []string { return []string{"foo", "bar"} })),
		That("f").Puts("foo", "bar").
			WithSetup(f(func() [2]string { return [2]string{"foo", "bar"} })),
		// Named types with underlying slice type treated as a single value
		That("f").Puts(namedSlice{"foo", "bar"}).
			WithSetup(f(func() namedSlice { return namedSlice{"foo", "bar"} })),

		// Error return value
		That("f").Throws(anError).
			WithSetup(f(func() (string, error) { return "x", anError })),
		That("f").DoesNothing().
			WithSetup(f(func() error { return nil })),
	)
}

func f(body any) func(*Evaler) {
	return func(ev *Evaler) {
		ev.ExtendGlobal(BuildNs().AddGoFn("f", body))
	}
}

func testInputs(t *testing.T, wantValues ...any) func(Inputs) {
	return func(i Inputs) {
		t.Helper()
		var values []any
		i(func(x any) {
			values = append(values, x)
		})
		wantValues := []any{"foo", "bar"}
		if !reflect.DeepEqual(values, wantValues) {
			t.Errorf("Inputs parameter didn't get supplied inputs")
		}
	}
}
