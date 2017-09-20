package eval

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elves/elvish/util"
)

var mods = map[string]string{
	"lorem":    "name = lorem",
	"d":        "name = d",
	"a/b/c/d":  "name = a/b/c/d",
	"has/init": "put has/init",
}

var dataDir string

func TestMain(m *testing.M) {
	var code int
	util.WithTempDir(func(dir string) {
		dataDir = dir
		for mod, content := range mods {
			fname := filepath.Join(dir, "lib", mod+".elv")
			os.MkdirAll(filepath.Dir(fname), 0755)
			f, err := os.Create(fname)
			if err != nil {
				panic(err)
			}
			_, err = f.WriteString(content)
			if err != nil {
				panic(err)
			}
			err = f.Close()
			if err != nil {
				panic(err)
			}
		}
		code = m.Run()
	})
	os.Exit(code)
}
