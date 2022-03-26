package fsutil

import (
	"fmt"
	"path/filepath"
	"sort"
	"testing"

	"src.elv.sh/pkg/testutil"
)

var claimFileTests = []struct {
	dir          string
	pattern      string
	wantFileName string
}{
	{".", "a*.log", "a9.log"},
	{".", "*.txt", "1.txt"},
	{"d", "*.txt", filepath.Join("d", "1.txt")},
}

func TestClaimFile(t *testing.T) {
	testutil.InTempDir(t)

	testutil.ApplyDir(testutil.Dir{
		"a0.log": "",
		"a1.log": "",
		"a8.log": "",
		"d":      testutil.Dir{}})

	for _, test := range claimFileTests {
		name := claimFileAndGetName(test.dir, test.pattern)
		if name != test.wantFileName {
			t.Errorf("ClaimFile claims %s, want %s", name, test.wantFileName)
		}
	}
}

func TestClaimFile_Concurrent(t *testing.T) {
	testutil.InTempDir(t)

	n := 9
	ch := make(chan string, n)
	for i := 0; i < n; i++ {
		go func() {
			ch <- claimFileAndGetName(".", "a*.log")
		}()
	}

	names := make([]string, n)
	for i := 0; i < n; i++ {
		names[i] = <-ch
	}
	sort.Strings(names)

	for i, name := range names {
		wantName := fmt.Sprintf("a%d.log", i+1)
		if name != wantName {
			t.Errorf("got names[%d] = %q, want %q", i, name, wantName)
		}
	}
}

func claimFileAndGetName(dir, pattern string) string {
	f, err := ClaimFile(dir, pattern)
	if err != nil {
		panic(fmt.Sprintf("ClaimFile errors: %v", err))
	}
	defer f.Close()
	return f.Name()
}
