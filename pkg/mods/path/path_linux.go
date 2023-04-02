//go:build linux

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
		// The coercion to int64 is so this works on 32-bit Linux platforms. On
		// 64-bit Linux platforms the typecast is a no-op.
		ATime:      time.Unix(int64(extra.Atim.Sec), int64(extra.Atim.Nsec)),
		CTime:      time.Unix(int64(extra.Ctim.Sec), int64(extra.Ctim.Nsec)),
		Inode:      new(big.Int).SetUint64(extra.Ino),
		Uid:        new(big.Int).SetUint64(uint64(extra.Uid)),
		Gid:        new(big.Int).SetUint64(uint64(extra.Gid)),
		NumLinks:   new(big.Int).SetUint64(uint64(extra.Nlink)),
		Device:     new(big.Int).SetUint64(uint64(extra.Dev)),
		RawDevice:  new(big.Int).SetUint64(uint64(extra.Rdev)),
		BlockSize:  new(big.Int).SetUint64(uint64(extra.Blksize)),
		BlockCount: new(big.Int).SetUint64(uint64(extra.Blocks)),
		Owner:      ownerName,
		Group:      groupName,
	}
}
