package lscolors

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"testing"

	"src.elv.sh/pkg/testutil"
)

type opt struct {
	setupErr error
	mh       bool
	wantErr  bool
}

func TestDetermineFeature(t *testing.T) {
	testutil.InTempDir(t)

	test := func(name, fname string, wantFeature feature, o opt) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			t.Helper()
			if o.setupErr != nil {
				t.Skip("skipped due to setup error:", o.setupErr)
			}
			feature, err := determineFeature(fname, o.mh)
			wantErr := "nil"
			if o.wantErr {
				wantErr = "non-nil"
			}
			if (err != nil) != o.wantErr {
				t.Errorf("determineFeature(%q, %v) returns error %v, want %v",
					fname, o.mh, err, wantErr)
			}
			if feature != wantFeature {
				t.Errorf("determineFeature(%q, %v) returns feature %v, want %v",
					fname, o.mh, feature, wantFeature)
			}
		})
	}

	err := create("a")
	test("regular file", "a", featureRegular, opt{setupErr: err})

	err = os.Symlink("a", "l")
	test("symlink", "l", featureSymlink, opt{setupErr: err})

	err = os.Symlink("aaaa", "lbad")
	test("broken symlink", "lbad", featureOrphanedSymlink, opt{setupErr: err})

	if runtime.GOOS != "windows" {
		err := os.Link("a", "a2")
		test("multi hard link", "a", featureMultiHardLink, opt{mh: true, setupErr: err})
		test("ignoring multi hard link", "a", featureRegular, opt{setupErr: err})
	}

	err = createNamedPipe("fifo")
	test("named pipe", "fifo", featureNamedPipe, opt{setupErr: err})

	l, err := net.Listen("unix", "sock")
	if err == nil {
		defer l.Close()
	}
	test("socket", "sock", featureSocket, opt{setupErr: err})

	// TODO: Test featureDoor on Solaris

	chr, err := findDevice(os.ModeDevice | os.ModeCharDevice)
	test("char device", chr, featureCharDevice, opt{setupErr: err})

	blk, err := findDevice(os.ModeDevice)
	test("block device", blk, featureBlockDevice, opt{setupErr: err})

	err = createMode("xu", 0100)
	test("executable by user", "xu", featureExecutable, opt{setupErr: err})
	err = createMode("xg", 0010)
	test("executable by group", "xg", featureExecutable, opt{setupErr: err})
	err = createMode("xo", 0001)
	test("executable by other", "xo", featureExecutable, opt{setupErr: err})

	err = createMode("su", 0600|os.ModeSetuid)
	test("setuid", "su", featureSetuid, opt{setupErr: err})
	err = createMode("sg", 0600|os.ModeSetgid)
	test("setgid", "sg", featureSetgid, opt{setupErr: err})

	test("nonexistent file", "nonexistent", featureInvalid, opt{wantErr: true})
}

func create(fname string) error {
	f, err := os.Create(fname)
	if err == nil {
		f.Close()
	}
	return err
}

func createMode(fname string, mode os.FileMode) error {
	f, err := os.OpenFile(fname, os.O_CREATE, mode)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Mode() != mode {
		return fmt.Errorf("created file has mode %v, want %v", info.Mode(), mode)
	}
	return nil
}

func findDevice(typ os.FileMode) (string, error) {
	entries, err := os.ReadDir("/dev")
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.Type() == typ {
			return "/dev/" + entry.Name(), nil
		}
	}
	return "", fmt.Errorf("can't find %v device under /dev", typ)
}
