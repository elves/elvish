package web

import (
	"testing"

	"src.elv.sh/pkg/prog"
	. "src.elv.sh/pkg/prog/progtest"
)

func TestWeb_SpuriousArgument(t *testing.T) {
	f := Setup(t)

	exit := prog.Run(f.Fds(), Elvish("-web", "x"), Program)

	TestError(t, f, exit, "arguments are not allowed with -web")
}

func TestWeb_SpuriousC(t *testing.T) {
	f := Setup(t)

	exit := prog.Run(f.Fds(), Elvish("-web", "-c"), Program)

	TestError(t, f, exit, "-c cannot be used together with -web")
}
