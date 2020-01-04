package eval

import "testing"

func TestMathModule(t *testing.T) {
	TestWithSetup(t,
		func(ev *Evaler) {
			ev.InstallModule("math", MathNs)
			err := ev.EvalSourceInTTY(NewInteractiveSource(`use math`))
			if err != nil {
				panic(err)
			}
		},

		That(`math:abs 2.1`).Puts(2.1),
		That(`math:abs -2.1`).Puts(2.1),

		That(`math:ceil 2.1`).Puts(3.0),
		That(`math:ceil -2.1`).Puts(-2.0),

		That(`math:floor 2.1`).Puts(2.0),
		That(`math:floor -2.1`).Puts(-3.0),

		That(`math:round 2.1`).Puts(2.0),
		That(`math:round -2.1`).Puts(-2.0),
		That(`math:round 2.5`).Puts(3.0),
		That(`math:round -2.5`).Puts(-3.0),
	)
}
