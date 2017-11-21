package lscolors

import (
	"os"
	"syscall"
)

//go:generate stringer -type=feature -output=feature_string.go

type feature int

const (
	featureInvalid feature = iota

	featureOrphanedSymlink
	featureSymlink

	featureMultiHardLink

	featureNamedPipe
	featureSocket
	featureDoor
	featureBlockDevice
	featureCharDevice

	featureWorldWritableStickyDirectory
	featureWorldWritableDirectory
	featureStickyDirectory
	featureDirectory

	featureCapability

	featureSetuid
	featureSetgid
	featureExecutable

	featureRegular
)

// Weirdly, permission masks for group and other are missing on platforms other
// than linux, darwin and netbsd. So we replicate some of them here.
const (
	_S_IWOTH = 0x2 // Writable by other
	_S_IXGRP = 0x8 // Executable by group
	_S_IXOTH = 0x1 // Executable by other
)

func determineFeature(fname string, mh bool) (feature, error) {
	var stat syscall.Stat_t
	err := syscall.Lstat(fname, &stat)
	if err != nil {
		return 0, err
	}

	// The type of syscall.Stat_t.Mode is uint32 on Linux and uint16 on Mac
	m := (uint32)(stat.Mode)

	// Symlink and OrphanedSymlink has highest precedence
	if is(m, syscall.S_IFLNK) {
		_, err := os.Stat(fname)
		if err != nil {
			return featureOrphanedSymlink, nil
		}
		return featureSymlink, nil
	}

	// featureMultiHardLink
	if mh && stat.Nlink > 1 {
		return featureMultiHardLink, nil
	}

	// type bits features
	switch {
	case is(m, syscall.S_IFIFO):
		return featureNamedPipe, nil
	case is(m, syscall.S_IFSOCK):
		return featureSocket, nil
		/*
			case m | syscall.S_IFDOOR != 0:
				return featureDoor, nil
		*/
	case is(m, syscall.S_IFBLK):
		return featureBlockDevice, nil
	case is(m, syscall.S_IFCHR):
		return featureCharDevice, nil
	case is(m, syscall.S_IFDIR):
		// Perm bits features for directory
		switch {
		case is(m, _S_IWOTH|syscall.S_ISVTX):
			return featureWorldWritableStickyDirectory, nil
		case is(m, _S_IWOTH):
			return featureWorldWritableDirectory, nil
		case is(m, syscall.S_ISVTX):
			return featureStickyDirectory, nil
		default:
			return featureDirectory, nil
		}
	}

	// TODO(xiaq): Support featureCapacity

	// Perm bits features for regular files
	switch {
	case is(m, syscall.S_ISUID):
		return featureSetuid, nil
	case is(m, syscall.S_ISGID):
		return featureSetgid, nil
	case m&(syscall.S_IXUSR|_S_IXGRP|_S_IXOTH) != 0:
		return featureExecutable, nil
	}

	// Check extension
	return featureRegular, nil
}

func is(u, p uint32) bool {
	return u&p == p
}
