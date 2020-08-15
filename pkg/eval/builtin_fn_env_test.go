package eval

import (
	"os"
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/util"
)

func TestBuiltinFnEnv(t *testing.T) {
	oldpath := os.Getenv(util.EnvPATH)
	listSep := string(os.PathListSeparator)
	Test(t,
		That(`get-env var`).ThrowsCause(errNonExistentEnvVar),
		That(`set-env var test1`),
		That(`get-env var`).Puts("test1"),
		That(`put $E:var`).Puts("test1"),

		That(`set-env var test2`),
		That(`get-env var`).Puts("test2"),
		That(`put $E:var`).Puts("test2"),

		That(`has-env var`).Puts(true),
		That(`unset-env var`),
		That(`has-env var`).Puts(false),

		That(`set-env PATH /test-path`),
		That(`put $paths`).Puts(vals.MakeList("/test-path")),
		That(`paths = [/test-path2 $@paths]`),
		That(`get-env PATH`).Puts("/test-path2"+listSep+"/test-path"),
	)
	os.Setenv(util.EnvPATH, oldpath)
}
