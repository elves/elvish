// Package eval handles evaluation of nodes and consists the runtime of the
// shell.
package eval

import (
	"os"
	"strings"
)

var env map[string]string
var search_paths []string

func envAsMap(env []string) (m map[string]string) {
	m = make(map[string]string)
	for _, e := range env {
		arr := strings.SplitN(e, "=", 2)
		if len(arr) == 2 {
			m[arr[0]] = arr[1]
		}
	}
	return
}

func init() {
	env = envAsMap(os.Environ())

	path_var, ok := env["PATH"]
	if ok {
		search_paths = strings.Split(path_var, ":")
		// fmt.Printf("Search paths are %v\n", search_paths)
	} else {
		search_paths = []string{"/bin"}
	}
}
