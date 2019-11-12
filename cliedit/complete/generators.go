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

// GenerateFileNames returns filename candidates that are suitable for completing
// the last argument. It can be used in Config.ArgGenerator.
func GenerateFileNames(args []string) ([]RawItem, error) {
	return generateFileNames(args[len(args)-1], false)
}

// GenerateForSudo generates candidates for sudo.
func GenerateForSudo(cfg Config, args []string) ([]RawItem, error) {
	switch {
	case len(args) < 2:
		return nil, errNoCompletion
	case len(args) == 2:
		// Complete external commands.
		return generateExternalCommands(args[1], cfg.PureEvaler)
	default:
		return cfg.ArgGenerator(args[1:])
	}
}

// Internal generators, used from completers.

func generateExternalCommands(seed string, ev PureEvaler) ([]RawItem, error) {
	if util.DontSearch(seed) {
		// Completing a local external command name.
		return generateFileNames(seed, true)
	}
	var items []RawItem
	ev.EachExternal(func(s string) { items = append(items, PlainItem(s)) })
	return items, nil
}

func generateCommands(seed string, ev PureEvaler) ([]RawItem, error) {
	if util.DontSearch(seed) {
		// Completing a local external command name.
		return generateFileNames(seed, true)
	}

	var cands []RawItem
	addPlainItem := func(s string) { cands = append(cands, PlainItem(s)) }

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

	sigil, qname := eval.SplitVariableRef(seed)
	ns, _ := eval.SplitQNameNsIncomplete(qname)
	if sigil == "" {
		// Generate functions, namespaces, and variable assignments.
		ev.EachVariableInNs(ns, func(varname string) {
			switch {
			case strings.HasSuffix(varname, eval.FnSuffix):
				addPlainItem(
					ns + varname[:len(varname)-len(eval.FnSuffix)])
			case strings.HasSuffix(varname, eval.NsSuffix):
				addPlainItem(ns + varname)
			default:
				cands = append(cands, noQuoteItem(ns+varname+" = "))
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

		items = append(items, ComplexItem{
			Stem: full, CodeSuffix: suffix,
			// style: ui.StylesFromString(lsColor.GetStyle(full)),
		})
	}

	return items, nil
}

func generateIndicies(v interface{}) []RawItem {
	var items []RawItem
	vals.IterateKeys(v, func(k interface{}) bool {
		if kstring, ok := k.(string); ok {
			items = append(items, PlainItem(kstring))
		}
		return true
	})
	return items
}

func dotfile(fname string) bool {
	return strings.HasPrefix(fname, ".")
}
