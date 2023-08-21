package os

import (
	"syscall"

	"src.elv.sh/pkg/eval/vals"
)

// Taken from
// https://learn.microsoft.com/en-us/windows/win32/fileio/file-attribute-constants.
// The syscall package only has a subset of these.
//
// Some of these attributes are redundant with fields in the outer stat map, but
// we keep all of them for consistency.
var fileAttributeNames = [...]struct {
	bit  uint32
	name string
}{
	{0x1, "readonly"},
	{0x2, "hidden"},
	{0x4, "system"},
	{0x10, "directory"},
	{0x20, "archive"},
	{0x40, "device"},
	{0x80, "normal"},
	{0x100, "temporary"},
	{0x200, "sparse-file"},
	{0x400, "reparse-point"},
	{0x800, "compressed"},
	{0x1000, "offline"},
	{0x2000, "not-content-indexed"},
	{0x4000, "encrypted"},
	{0x8000, "integrity-system"},
	{0x10000, "virtual"},
	{0x20000, "no-scrub-data"},
	{0x40000, "ea"},
	{0x80000, "pinned"},
	{0x100000, "unpinned"},
	{0x400000, "recall-on-data-access"},
}

func statSysMap(sys any) vals.Map {
	attrData := sys.(*syscall.Win32FileAttributeData)
	// TODO: Make this a set when Elvish has a set type.
	fileAttributes := vals.EmptyList
	for _, attr := range fileAttributeNames {
		if attrData.FileAttributes&attr.bit != 0 {
			fileAttributes = fileAttributes.Conj(attr.name)
		}
	}
	return vals.MakeMap(
		"file-attributes", fileAttributes,
		// TODO: CreationTime, LastAccessTime, LastWriteTime
	)
}
