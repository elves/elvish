package web

import (
	"testing"

	"github.com/elves/elvish/pkg/prog"
	. "github.com/elves/elvish/pkg/prog/progtest"
)

func TestWeb_SpuriousArgument(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := prog.Run(f.Fds(), Elvish("-web", "x"), Program)

	TestError(t, f, exit, "arguments are not allowed with -web")
}

func TestWeb_SpuriousC(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := prog.Run(f.Fds(), Elvish("-web", "-c"), Program)

	TestError(t, f, exit, "-c cannot be used together with -web")
}
