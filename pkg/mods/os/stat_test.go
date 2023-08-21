package os_test

import (
	"os"
	"testing"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/must"
)

func TestStat(t *testing.T) {
	InTempDir(t)
	ApplyDir(Dir{
		"dir":  Dir{},
		"file": "foobar",
	})

	TestWithEvalerSetup(t, useOS,
		That(`os:stat file`).Puts(MapContainingPairs(
			"name", "file",
			"size", 6,
			"type", "regular",
		)),
		That(`os:stat dir`).Puts(MapContainingPairs(
			"name", "dir",
			// size field of directories is platform-dependent
			"type", "dir",
		)),
		That(`os:stat non-existent`).Throws(ErrorWithType(&os.PathError{})),
	)
}

func TestStat_Symlink(t *testing.T) {
	InTempDir(t)
	ApplyDir(Dir{"regular": ""})
	err := os.Symlink("regular", "symlink")
	if err != nil {
		// On Windows we may or may not be able to create a symlink.
		t.Skipf("symlink: %v", err)
	}

	TestWithEvalerSetup(t, useOS,
		That(`os:stat symlink`).
			Puts(MapContainingPairs("type", "symlink")),
		That(`os:stat &follow-symlink=$true symlink`).
			Puts(MapContainingPairs("type", "regular")),
	)
}

var permAndSpecialModesTests = []struct {
	name    string
	mode    os.FileMode
	statMap vals.Map
}{
	{"444", 0o444, vals.MakeMap("perm", 0o444)},
	{"666", 0o666, vals.MakeMap("perm", 0o666)},
	{"setuid", os.ModeSetuid, vals.MakeMap("special-modes", vals.MakeList("setuid"))},
	{"setgid", os.ModeSetgid, vals.MakeMap("special-modes", vals.MakeList("setgid"))},
	{"sticky", os.ModeSticky, vals.MakeMap("special-modes", vals.MakeList("sticky"))},
}

func TestStat_PermAndSpecialModes(t *testing.T) {
	Umask(t, 0)
	for _, test := range permAndSpecialModesTests {
		t.Run(test.name, func(t *testing.T) {
			InTempDir(t)
			must.OK(os.WriteFile("file", nil, 0o666))
			ChmodOrSkip(t, "file", test.mode)

			TestWithEvalerSetup(t, useOS,
				That(`os:stat file`).Puts(MapContaining(test.statMap)),
			)
		})
	}
}
