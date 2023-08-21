package os_test

import (
	"src.elv.sh/pkg/eval/evaltest"
	osmod "src.elv.sh/pkg/mods/os"
	"src.elv.sh/pkg/testutil"
)

var (
	Test                = evaltest.Test
	TestWithEvalerSetup = evaltest.TestWithEvalerSetup
	TestWithSetup       = evaltest.TestWithSetup
	Use                 = evaltest.Use
	That                = evaltest.That
	AnyInteger          = evaltest.AnyInteger
	StringMatching      = evaltest.StringMatching
	MapContaining       = evaltest.MapContaining
	MapContainingPairs  = evaltest.MapContainingPairs
	ErrorWithType       = evaltest.ErrorWithType
	ErrorWithMessage    = evaltest.ErrorWithMessage
)

var (
	ApplyDir    = testutil.ApplyDir
	ChmodOrSkip = testutil.ChmodOrSkip
	InTempDir   = testutil.InTempDir
	Umask       = testutil.Umask
)

type Dir = testutil.Dir

var useOS = Use("os", osmod.Ns)
