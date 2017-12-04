package eval

import "io/ioutil"

// EachExternal calls f for each name that can resolve to an external
// command.
// TODO(xiaq): Windows support
func (ev *Evaler) EachExternal(f func(string)) {
	for _, dir := range ev.searchPaths() {
		// XXX Ignore error
		infos, _ := ioutil.ReadDir(dir)
		for _, info := range infos {
			if !info.IsDir() && (info.Mode()&0111 != 0) {
				f(info.Name())
			}
		}
	}
}
