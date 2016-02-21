package eval

import (
	"fmt"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// Search tries to resolve an external command and return the full (possibly
// relative) path.
func (ev *Evaler) Search(exe string) (string, error) {
	path, err := util.Search(ev.searchPaths(), exe)
	if err != nil {
		return "", fmt.Errorf("search %s: %s", parse.Quote(exe), err.Error())
	}
	return path, nil
}

// AllExecutables writes the names of all executable files in the search path
// to a channel.
func (ev *Evaler) AllExecutables(names chan<- string) {
	util.AllExecutables(ev.searchPaths(), names)
}
