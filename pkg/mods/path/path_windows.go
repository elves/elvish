//go:build windows

package path

import (
	_ "embed"
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
	return fileInfo{
		Path:         path,
		AbsPath:      absPath,
		IsDir:        info.IsDir(),
		Size:         big.NewInt(info.Size()),
		Mode:         new(big.Int).SetUint64(uint64(info.Mode())),
		SymbolicMode: info.Mode().String(),
		MTime:        info.ModTime(),
		ATime:        time.Unix(0, extra.LastAccessTime.Nanoseconds()),
		BTime:        time.Unix(0, extra.CreationTime.Nanoseconds()),
	}
}
