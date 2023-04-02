//go:build freebsd || netbsd

package path

import (
	_ "embed"
	"fmt"
	"io/fs"
	"math/big"
	"os/user"
	"path/filepath"
	"syscall"
	"time"
)

func pathMetadata(path string, info fs.FileInfo) fileInfo {
	extra := info.Sys().(*syscall.Stat_t)
	var groupName, ownerName string
	if user, err := user.LookupId(fmt.Sprintf("%d", extra.Uid)); err == nil {
		ownerName = user.Username
	} else {
		ownerName = "<unknown>"
	}
	if group, err := user.LookupGroupId(fmt.Sprintf("%d", extra.Gid)); err == nil {
		groupName = group.Name
	} else {
		groupName = "<unknown>"
	}
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
		ATime:        time.Unix(extra.Atimespec.Sec, extra.Atimespec.Nsec),
		BTime:        time.Unix(extra.Birthtimespec.Sec, extra.Birthtimespec.Nsec),
		CTime:        time.Unix(extra.Ctimespec.Sec, extra.Ctimespec.Nsec),
		Inode:        new(big.Int).SetUint64(extra.Ino),
		Uid:          new(big.Int).SetUint64(uint64(extra.Uid)),
		Gid:          new(big.Int).SetUint64(uint64(extra.Gid)),
		NumLinks:     new(big.Int).SetUint64(uint64(extra.Nlink)),
		Device:       new(big.Int).SetUint64(uint64(extra.Dev)),
		RawDevice:    new(big.Int).SetUint64(uint64(extra.Rdev)),
		BlockSize:    new(big.Int).SetUint64(uint64(extra.Blksize)),
		BlockCount:   new(big.Int).SetUint64(uint64(extra.Blocks)),
		Owner:        ownerName,
		Group:        groupName,
	}
}
