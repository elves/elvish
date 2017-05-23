package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGetwd(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "elvishtest.")
	if err != nil {
		panic(err)
	}
	pwd, err := os.Getwd()
	if err != nil {
		os.Chdir(pwd)
	}
	os.Chdir(tmpdir)

	// On some systems /tmp is a symlink.
	tmpdir, err = filepath.EvalSymlinks(tmpdir)
	if err != nil {
		panic(err)
	}
	if gotwd := Getwd(); gotwd != tmpdir {
		t.Errorf("Getwd() -> %v, want %v", gotwd, tmpdir)
	}

	// Override $HOME to trick GetHome.
	os.Setenv("HOME", tmpdir)

	if gotwd := Getwd(); gotwd != "~" {
		t.Errorf("Getwd() -> %v, want ~", gotwd)
	}

	mustOK(os.Mkdir("a", 0700))
	mustOK(os.Chdir("a"))
	if gotwd := Getwd(); gotwd != "~/a" {
		t.Errorf("Getwd() -> %v, want ~/a", gotwd)
	}

	mustOK(os.Remove(tmpdir + "/a"))
	mustOK(os.Remove(tmpdir))
	// XXX On OS X os.Getwd will still return the old path name in face of
	// directory being removed. We disable this test.
	/*
		if gotwd := Getwd(); gotwd != "?" {
			t.Errorf("Getwd() -> %v, want ?", gotwd)
		}
	*/
}

func mustOK(err error) {
	if err != nil {
		panic(err)
	}
}
