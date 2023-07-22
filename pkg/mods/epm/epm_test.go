package epm_test

import (
	"testing"

	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/mods"
)

func TestEPM(t *testing.T) {
	// A smoke test to ensure that the epm module has no errors.

	TestWithEvalerSetup(t, mods.AddTo,
		That("use epm").DoesNothing(),
	)
}
