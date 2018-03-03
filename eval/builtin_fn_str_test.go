package eval

import "testing"

func TestBuiltinFnStr(t *testing.T) {
	runTests(t, []Test{
		That(`==s haha haha`).Puts(true),
		That(`==s 10 10.0`).Puts(false),
		That(`<s a b`).Puts(true),
		That(`<s 2 10`).Puts(false),

		That(`joins : [/usr /bin /tmp]`).Puts("/usr:/bin:/tmp"),
		That(`joins : ['' a '']`).Puts(":a:"),
		That(`splits : /usr:/bin:/tmp`).Puts("/usr", "/bin", "/tmp"),
		That(`splits : /usr:/bin:/tmp &max=2`).Puts("/usr", "/bin:/tmp"),
		That(`replaces : / ":usr:bin:tmp"`).Puts("/usr/bin/tmp"),
		That(`replaces &max=2 : / :usr:bin:tmp`).Puts("/usr/bin:tmp"),

		That(`ord a`).Puts("0x61"),
		That(`base 16 42 233`).Puts("2a", "e9"),
		That(`base 1 1`).Errors(),   // no base-1
		That(`base 37 10`).Errors(), // no letter for base-37
		That(`wcswidth 你好`).Puts("4"),
		That(`-override-wcwidth x 10; wcswidth 1x2x; -override-wcwidth x 1`).Puts("22"),

		That(`has-prefix golang go`).Puts(true),
		That(`has-prefix golang x`).Puts(false),
		That(`has-suffix golang x`).Puts(false),

		That(`echo "  ax  by cz  \n11\t22 33" | eawk [@a]{ put $a[-1] }`).Puts(
			"cz", "33"),
	})
}
