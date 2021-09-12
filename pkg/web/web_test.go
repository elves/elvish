package web

import (
	"testing"

	. "src.elv.sh/pkg/prog/progtest"
)

func TestProgram(t *testing.T) {
	Test(t, Program,
		ThatElvish("-web", "x").
			ExitsWith(2).
			WritesStderrContaining("arguments are not allowed with -web"),

		ThatElvish("-web", "-c").
			ExitsWith(2).
			WritesStderrContaining("-c cannot be used together with -web"),
	)
}
