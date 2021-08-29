package fsutil

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ErrClaimFileBadPattern is thrown when the pattern argument passed to
// ClaimFile does not contain exactly one asterisk.
var ErrClaimFileBadPattern = errors.New("ClaimFile: pattern must contain exactly one asterisk")

// ClaimFile takes a directory and a pattern string containing exactly one
// asterisk (e.g. "a*.log"). It opens a file in that directory, with a filename
// matching the template, with "*" replaced by a number. That number is one plus
// the largest of all existing files matching the template. If no such file
// exists, "*" is replaced by 1. The file is opened for read and write, with
// permission 0666 (before umask).
//
// For example, if the directory /tmp/elvish contains a1.log, a2.log and a9.log,
// calling ClaimFile("/tmp/elvish", "a*.log") will open a10.log. If the
// directory has no files matching the pattern, this same call will open a1.log.
//
// This function is useful for automatically determining unique names for log
// files. Unique filenames can also be derived by embedding the PID, but using
// this function preserves the chronical order of the files.
//
// This function is concurrency-safe: it always opens a new, unclaimed file and
// is not subject to race condition.
func ClaimFile(dir, pattern string) (*os.File, error) {
	if strings.Count(pattern, "*") != 1 {
		return nil, ErrClaimFileBadPattern
	}
	asterisk := strings.IndexByte(pattern, '*')
	prefix, suffix := pattern[:asterisk], pattern[asterisk+1:]
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	max := 0
	for _, file := range files {
		name := file.Name()
		if len(name) > len(prefix)+len(suffix) && strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix) {
			core := name[len(prefix) : len(name)-len(suffix)]
			if coreNum, err := strconv.Atoi(core); err == nil {
				if max < coreNum {
					max = coreNum
				}
			}
		}
	}

	for i := max + 1; ; i++ {
		name := filepath.Join(dir, prefix+strconv.Itoa(i)+suffix)
		f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err == nil {
			return f, nil
		}
		if !os.IsExist(err) {
			return nil, err
		}
	}
}
