//go:build windows

package path

import (
	"io/fs"
	"math/big"
	"path/filepath"
	"syscall"
	"time"
)

const devTty = "CON"

func pathMetadata(path string, info fs.FileInfo) fileInfo {
	extra := info.Sys().(*syscall.Win32FileAttributeData)
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	perms, symbolicPerms := publicPerms(info)

	// Some of the mode translations (such as setting the IsNamedPipe struct
	// member) are probably never true on Windows. Nonetheless, we include them
	// for symmetry with Unix and because I might be mistaken in about whether
	// they are, or are not, applicable on Windows.
	return fileInfo{
		Path:        path,
		AbsPath:     absPath,
		IsDir:       info.IsDir(),
		IsRegular:   info.Mode().IsRegular(),
		IsNamedPipe: info.Mode()&fs.ModeNamedPipe == fs.ModeNamedPipe,
		IsSymlink:   info.Mode()&fs.ModeSymlink == fs.ModeSymlink,
		IsDevice:    info.Mode()&fs.ModeDevice == fs.ModeDevice,
		IsCharDevice: info.Mode()&fs.ModeDevice == fs.ModeDevice &&
			info.Mode()&fs.ModeCharDevice == fs.ModeCharDevice,
		Size:          big.NewInt(info.Size()),
		Mode:          new(big.Int).SetUint64(uint64(info.Mode())),
		SymbolicMode:  info.Mode().String(),
		Perms:         new(big.Int).SetUint64(perms),
		SymbolicPerms: symbolicPerms,
		MTime:         info.ModTime(),
		ATime:         time.Unix(0, extra.LastAccessTime.Nanoseconds()),
		BTime:         time.Unix(0, extra.CreationTime.Nanoseconds()),
	}
}
