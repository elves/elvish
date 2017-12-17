package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWithTempDirs_PassesDirs(t *testing.T) {
	WithTempDirs(10, func(dirs []string) {
		for _, dir := range dirs {
			stat, err := os.Stat(dir)
			if err != nil {
				t.Errorf("WithTempDir passes %q, but it cannot be stat'ed", dir)
			}
			if !stat.IsDir() {
				t.Errorf("WithTempDir passes %q, but it is not dir", dir)
			}
		}
	})
}

func TestWithTempDir_RemovesDirs(t *testing.T) {
	var tempDirs []string
	WithTempDirs(10, func(dirs []string) { tempDirs = dirs })
	for _, dir := range tempDirs {
		_, err := os.Stat(dir)
		if err == nil {
			t.Errorf("After WithTempDir returns, %q still exists", dir)
		}
	}
}

func TestInTempDir_CDIn(t *testing.T) {
	InTempDir(func(tmpDir string) {
		pwd := getPwd()
		evaledTmpDir, err := filepath.EvalSymlinks(tmpDir)
		if err != nil {
			panic(err)
		}
		if pwd != evaledTmpDir {
			t.Errorf("In InTempDir, working dir (%q) != EvalSymlinks(argument) (%q)", pwd, evaledTmpDir)
		}
	})
}

func TestInTempDir_CDOut(t *testing.T) {
	before := getPwd()
	InTempDir(func(tmpDir string) {})
	after := getPwd()
	if before != after {
		t.Errorf("With InTempDir, working dir before %q != after %q", before, after)
	}
}

func getPwd() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		panic(err)
	}
	return dir
}
