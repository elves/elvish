package lscolors

import (
	"os"
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
	worldWritable = 02   // Writable by other
	executable    = 0111 // Executable
)

func determineFeature(fname string, mh bool) (feature, error) {
	stat, err := os.Lstat(fname)

	if err != nil {
		return featureInvalid, err
	}

	m := stat.Mode()

	// Symlink and OrphanedSymlink has highest precedence
	if is(m, os.ModeSymlink) {
		_, err := os.Stat(fname)
		if err != nil {
			return featureOrphanedSymlink, nil
		}
		return featureSymlink, nil
	}

	// featureMultiHardLink
	if mh && isMultiHardlink(stat) {
		return featureMultiHardLink, nil
	}

	// type bits features
	switch {
	case is(m, os.ModeNamedPipe):
		return featureNamedPipe, nil
	case is(m, os.ModeSocket): // Never on Windows
		return featureSocket, nil
	case isDoor(stat):
		return featureDoor, nil
	case is(m, os.ModeCharDevice):
		return featureCharDevice, nil
	case is(m, os.ModeDevice):
		// There is no dedicated os.Mode* flag for block device. On all
		// supported Unix platforms, when os.ModeDevice is set but
		// os.ModeCharDevice is not, the file is a block device (i.e.
		// syscall.S_IFBLK is set). On Windows, this branch is unreachable.
		//
		// On Plan9, this in inaccurate.
		return featureBlockDevice, nil
	case is(m, os.ModeDir):
		// Perm bits features for directory
		perm := m.Perm()
		switch {
		case is(m, os.ModeSticky) && is(perm, worldWritable):
			return featureWorldWritableStickyDirectory, nil
		case is(perm, worldWritable):
			return featureWorldWritableDirectory, nil
		case is(m, os.ModeSticky):
			return featureStickyDirectory, nil
		default:
			return featureDirectory, nil
		}
	}

	// TODO(xiaq): Support featureCapacity

	// Perm bits features for regular files
	switch {
	case is(m, os.ModeSetuid):
		return featureSetuid, nil
	case is(m, os.ModeSetgid):
		return featureSetgid, nil
	case m&executable != 0:
		return featureExecutable, nil
	}

	// Check extension
	return featureRegular, nil
}

func is(m, p os.FileMode) bool {
	return m&p == p
}
