package eval_test

import (
	"os"
	"testing"

	"src.elv.sh/pkg/eval"

	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

func TestGetEnv(t *testing.T) {
	restore := saveEnv("var")
	defer restore()

	os.Unsetenv("var")
	Test(t, That(`get-env var`).Throws(eval.ErrNonExistentEnvVar))

	os.Setenv("var", "test1")
	Test(t,
		That(`get-env var`).Puts("test1"),
		That(`put $E:var`).Puts("test1"),
	)

	os.Setenv("var", "test2")
	Test(t,
		That(`get-env var`).Puts("test2"),
		That(`put $E:var`).Puts("test2"),
	)
}

func TestHasEnv(t *testing.T) {
	restore := saveEnv("var")
	defer restore()

	os.Setenv("var", "test1")
	Test(t, That(`has-env var`).Puts(true))

	os.Unsetenv("var")
	Test(t, That(`has-env var`).Puts(false))
}

func TestSetEnv(t *testing.T) {
	restore := saveEnv("var")
	defer restore()

	Test(t, That("set-env var test1").DoesNothing())
	if envVal := os.Getenv("var"); envVal != "test1" {
		t.Errorf("got $E:var = %q, want 'test1'", envVal)
	}
}

func TestSetEnv_PATH(t *testing.T) {
	restore := saveEnv("PATH")
	defer restore()

	listSep := string(os.PathListSeparator)
	Test(t,
		That(`set-env PATH /test-path`),
		That(`put $paths`).Puts(vals.MakeList("/test-path")),
		That(`set paths = [/test-path2 $@paths]`),
		That(`set paths = [$true]`).Throws(vars.ErrPathMustBeString),
		That(`set paths = ["/invalid`+string(os.PathListSeparator)+`:path"]`).
			Throws(vars.ErrPathContainsForbiddenChar),
		That(`set paths = ["/invalid\000path"]`).
			Throws(vars.ErrPathContainsForbiddenChar),
		That(`get-env PATH`).Puts("/test-path2"+listSep+"/test-path"),
	)
}

func saveEnv(name string) func() {
	oldValue, ok := os.LookupEnv(name)
	return func() {
		if ok {
			os.Setenv(name, oldValue)
		}
	}
}
