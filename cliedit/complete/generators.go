package complete

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

var quotedPathSeparator = parse.Quote(string(filepath.Separator))

func generateCommands(seed string, ev PureEvaler) ([]RawItem, error) {
	if util.DontSearch(seed) {
		// Completing a local external command name.
		return generateFileNames(seed, true)
	}

	var cands []RawItem
	addPlainItem := func(s string) { cands = append(cands, plainItem(s)) }

	if strings.HasPrefix(seed, "e:") {
		// Generate all external commands with the e: prefix, and be done.
		ev.EachExternal(func(command string) {
			addPlainItem("e:" + command)
		})
		return cands, nil
	}

	// Generate all special forms.
	ev.EachSpecial(addPlainItem)
	// Generate all external commands (without the e: prefix).
	ev.EachExternal(addPlainItem)

	explode, ns, _ := eval.ParseIncompleteVariableRef(seed)
	if !explode {
		// Generate functions, namespaces, and variable assignments.
		ev.EachVariableInNs(ns, func(varname string) {
			switch {
			case strings.HasSuffix(varname, eval.FnSuffix):
				addPlainItem(eval.MakeVariableRef(
					false, ns, varname[:len(varname)-len(eval.FnSuffix)]))
			case strings.HasSuffix(varname, eval.NsSuffix):
				addPlainItem(eval.MakeVariableRef(false, ns, varname))
			default:
				name := eval.MakeVariableRef(false, ns, varname)
				cands = append(cands, &complexItem{name, " = ", " = "})
			}
		})
	}

	return cands, nil
}

func generateFileNames(seed string, onlyExecutable bool) ([]RawItem, error) {
	var items []RawItem

	dir, fileprefix := filepath.Split(seed)
	dirToRead := dir
	if dirToRead == "" {
		dirToRead = "."
	}

	infos, err := ioutil.ReadDir(dirToRead)
	if err != nil {
		return nil, fmt.Errorf("cannot list directory %s: %v", dirToRead, err)
	}

	// lsColor := lscolors.GetColorist()

	// Make candidates out of elements that match the file component.
	for _, info := range infos {
		name := info.Name()
		// Show dot files iff file part of pattern starts with dot, and vice
		// versa.
		if dotfile(fileprefix) != dotfile(name) {
			continue
		}
		// Only accept searchable directories and executable files if
		// executableOnly is true.
		if onlyExecutable && (info.Mode()&0111) == 0 {
			continue
		}

		// Full filename for source and getStyle.
		full := dir + name

		suffix := " "
		if info.IsDir() {
			suffix = quotedPathSeparator
		} else if info.Mode()&os.ModeSymlink != 0 {
			stat, err := os.Stat(full)
			if err == nil && stat.IsDir() {
				// Symlink to directory.
				suffix = quotedPathSeparator
			}
		}

		items = append(items, &complexItem{
			stem: full, codeSuffix: suffix,
			// style: ui.StylesFromString(lsColor.GetStyle(full)),
		})
	}

	return items, nil
}

func generateIndicies(v interface{}) []RawItem {
	var items []RawItem
	vals.IterateKeys(v, func(k interface{}) bool {
		if kstring, ok := k.(string); ok {
			items = append(items, plainItem(kstring))
		}
		return true
	})
	return items
}

func dotfile(fname string) bool {
	return strings.HasPrefix(fname, ".")
}
