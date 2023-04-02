//go:build unix

package path

import (
	"fmt"
	"io/fs"
	"math/big"
	"os/user"
	"path/filepath"
	"syscall"
)

const devTty = "/dev/tty"

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
	perms, symbolicPerms := publicPerms(info)

	// The coercion to uint64 below is so this works on 32-bit platforms where
	// values like extra.Uid are 32-bit.
	fi := fileInfo{
		Path:        path,
		AbsPath:     absPath,
		IsDir:       info.IsDir(),
		IsRegular:   info.Mode().IsRegular(),
		IsNamedPipe: info.Mode()&fs.ModeNamedPipe == fs.ModeNamedPipe,
		IsSymlink:   info.Mode()&fs.ModeSymlink == fs.ModeSymlink,
		IsDevice:    info.Mode()&fs.ModeDevice == fs.ModeDevice,
		IsCharDevice: info.Mode()&fs.ModeDevice == fs.ModeDevice &&
			info.Mode()&fs.ModeCharDevice == fs.ModeCharDevice,
		Mode:          new(big.Int).SetUint64(uint64(info.Mode())),
		SymbolicMode:  info.Mode().String(),
		Perms:         new(big.Int).SetUint64(perms),
		SymbolicPerms: symbolicPerms,
		MTime:         info.ModTime(),
		Size:          big.NewInt(info.Size()),
		Inode:         new(big.Int).SetUint64(extra.Ino),
		Uid:           new(big.Int).SetUint64(uint64(extra.Uid)),
		Gid:           new(big.Int).SetUint64(uint64(extra.Gid)),
		NumLinks:      new(big.Int).SetUint64(uint64(extra.Nlink)),
		Device:        new(big.Int).SetUint64(uint64(extra.Dev)),
		RawDevice:     new(big.Int).SetUint64(uint64(extra.Rdev)),
		BlockSize:     new(big.Int).SetUint64(uint64(extra.Blksize)),
		BlockCount:    new(big.Int).SetUint64(uint64(extra.Blocks)),
		Owner:         ownerName,
		Group:         groupName,
	}
	// Apply any Unix Unix OS specific modifications of the structure. This is
	// just the auxiliary timestamps at this time.
	pathPlatformMetadata(&fi, extra)
	return fi
}
