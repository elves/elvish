package eval

import "testing"

func TestClosure(t *testing.T) {
	runTests(t, []Test{
		That("kind-of { }").Puts("fn"),
		That("eq { } { }").Puts(false),
		That("x = { }; put [&$x= foo][$x]").Puts("foo"),
		That("[x]{ } a b").ErrorsAny(),
		That("[x y]{ } a").ErrorsAny(),
		That("[x y @rest]{ } a").ErrorsAny(),
		That("[]{ } &k=v").ErrorsAny(),

		That("explode [a b]{ }[arg-names]").Puts("a", "b"),
		That("put [@r]{ }[rest-arg]").Puts("r"),
		That("explode [&opt=def]{ }[opt-names]").Puts("opt"),
		That("explode [&opt=def]{ }[opt-defaults]").Puts("def"),
		That("put { body }[src][code]").Puts(
			"put { body }[src][code]"),
		That("put { body }[body]").Puts(" body "),
	})
}
