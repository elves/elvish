package os_test

import (
	"testing"

	"golang.org/x/sys/windows"

	"src.elv.sh/pkg/must"
)

func TestStat_Sys_FileAttributes(t *testing.T) {
	InTempDir(t)
	ApplyDir(Dir{
		"directory": Dir{},
		"readonly":  "",
		"hidden":    "",
	})
	mustSetFileAttributes("readonly", windows.FILE_ATTRIBUTE_READONLY)
	mustSetFileAttributes("hidden", windows.FILE_ATTRIBUTE_HIDDEN)

	TestWithEvalerSetup(t, useOS,
		That(`has-value (os:stat directory)[sys][file-attributes] directory`).
			Puts(true),
		That(`has-value (os:stat readonly)[sys][file-attributes] readonly`).
			Puts(true),
		That(`has-value (os:stat hidden)[sys][file-attributes] hidden`).
			Puts(true),
	)

}

func mustSetFileAttributes(name string, attr uint32) {
	must.OK(windows.SetFileAttributes(must.OK1(windows.UTF16PtrFromString(name)), attr))
}
