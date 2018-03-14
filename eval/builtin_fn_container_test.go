package eval

import "testing"

func TestBuiltinFnContainer(t *testing.T) {
	runTests(t, []Test{
		That(`ns: = (ns [&a=b &x=y]); put $ns:a`).Puts("b"),
		That(`ns: = (ns [&a=b &x=y]); keys $ns:`).Puts("a", "x"),
		That(`ns: = (ns [&a=b &x=y]); has-key $ns: a`).Puts(true),
		That(`ns: = (ns [&a=b &x=y]); has-key $ns: b`).Puts(false),

		That(`range 3`).Puts("0", "1", "2"),
		That(`range 1 3`).Puts("1", "2"),
		That(`range 0 10 &step=3`).Puts("0", "3", "6", "9"),
		That(`repeat 4 foo`).Puts("foo", "foo", "foo", "foo"),
		That(`explode [foo bar]`).Puts("foo", "bar"),

		That(`put (assoc [0] 0 zero)[0]`).Puts("zero"),
		That(`put (assoc [&] k v)[k]`).Puts("v"),
		That(`put (assoc [&k=v] k v2)[k]`).Puts("v2"),
		That(`has-key (dissoc [&k=v] k) k`).Puts(false),

		That(`put foo bar | all`).Puts("foo", "bar"),
		That(`echo foobar | all`).Prints("foobar\n"),
		That(`{ put foo bar; echo foobar } | all`).Puts(
			"foo", "bar").Prints("foobar\n"),
		That(`range 100 | take 2`).Puts("0", "1"),
		That(`range 100 | drop 98`).Puts("98", "99"),

		That(`has-key [foo bar] 0`).Puts(true),
		That(`has-key [foo bar] 0:1`).Puts(true),
		That(`has-key [foo bar] 0:20`).Puts(false),
		That(`has-key [&lorem=ipsum &foo=bar] lorem`).Puts(true),
		That(`has-key [&lorem=ipsum &foo=bar] loremwsq`).Puts(false),
		That(`has-value [&lorem=ipsum &foo=bar] lorem`).Puts(false),
		That(`has-value [&lorem=ipsum &foo=bar] bar`).Puts(true),
		That(`has-value [foo bar] bar`).Puts(true),
		That(`has-value [foo bar] badehose`).Puts(false),
		That(`has-value "foo" o`).Puts(true),
		That(`has-value "foo" d`).Puts(false),

		That(`range 100 | count`).Puts("100"),
		That(`count [(range 100)]`).Puts("100"),
		That(`count 1 2 3`).Errors(),

		That(`keys [&]`).DoesNothing(),
		That(`keys [&a=foo]`).Puts("a"),
		// Windows does not have an external sort command. Disabled until we have a
		// builtin sort command.
		// That(`keys [&a=foo &b=bar] | each echo | sort | each $put~`).Puts("a", "b"),
	})
}
