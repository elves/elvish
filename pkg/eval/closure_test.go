package eval

import "testing"

func TestClosure(t *testing.T) {
	Test(t,
		That("kind-of { }").Puts("fn"),
		That("eq { } { }").Puts(false),
		That("x = { }; put [&$x= foo][$x]").Puts("foo"),
		That("[x]{ } a b").ThrowsAny(),
		That("[x y]{ } a").ThrowsAny(),
		That("[x y @rest]{ } a").ThrowsAny(),
		That("[]{ } &k=v").ThrowsAny(),

		That("explode [a b]{ }[arg-names]").Puts("a", "b"),
		That("put [@r]{ }[rest-arg]").Puts("r"),
		That("explode [&opt=def]{ }[opt-names]").Puts("opt"),
		That("explode [&opt=def]{ }[opt-defaults]").Puts("def"),
		That("put { body }[body]").Puts(" body "),
		That("put [x @y]{ body }[def]").Puts("[x @y]{ body }"),
		That("put { body }[src][code]").
			Puts("put { body }[src][code]"),
	)
}
