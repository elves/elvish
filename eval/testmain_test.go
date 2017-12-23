package eval

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/elves/elvish/util"
)

var (
	filesToCreate = []string{
		"a1", "a2", "a3", "a10",
		"b1", "b2", "b3",
		"c1", "c2",
		"foo", "bar", "lorem", "ipsum",
	}
	dirsToCreate = []string{"dir", "dir2"}
	fileListing  = getFileListing()
)

func getFileListing() []string {
	var x []string
	x = append(x, filesToCreate...)
	x = append(x, dirsToCreate...)
	sort.Strings(x)
	return x
}

func getFilesWithPrefix(prefixes ...string) []string {
	var x []string
	for _, name := range fileListing {
		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				x = append(x, name)
				break
			}
		}
	}
	sort.Strings(x)
	return x
}

func getFilesBut(excludes ...string) []string {
	var x []string
	for _, name := range fileListing {
		excluded := false
		for _, exclude := range excludes {
			if name == exclude {
				excluded = true
				break
			}
		}
		if !excluded {
			x = append(x, name)
		}
	}
	sort.Strings(x)
	return x
}

var mods = map[string]string{
	"lorem":    "name = lorem; fn put-name { put $name }",
	"d":        "name = d",
	"a/b/c/d":  "name = a/b/c/d",
	"a/b/c/x":  "use ./d; d = $d:name; use ../../../lorem; lorem = $lorem:name",
	"has/init": "put has/init",
}

var libDir string

func TestMain(m *testing.M) {
	var exitCode int
	util.InTempDir(func(tmpHome string) {
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpHome)
		defer os.Setenv("HOME", oldHome)

		for _, filename := range filesToCreate {
			file, err := os.Create(filename)
			if err != nil {
				panic(err)
			}
			file.Close()
		}

		for _, dirname := range dirsToCreate {
			err := os.Mkdir(dirname, 0700)
			if err != nil {
				panic(err)
			}
		}

		util.WithTempDir(func(dir string) {
			libDir = dir

			for mod, content := range mods {
				fname := filepath.Join(libDir, mod+".elv")
				os.MkdirAll(filepath.Dir(fname), 0700)
				err := ioutil.WriteFile(fname, []byte(content), 0600)
				if err != nil {
					panic(err)
				}
			}

			exitCode = m.Run()
		})
	})
	os.Exit(exitCode)
}

func runTests(t *testing.T, tests []Test) {
	RunTests(t, tests, func() *Evaler {
		ev := NewEvaler()
		ev.SetLibDir(libDir)
		return ev
	})
}
