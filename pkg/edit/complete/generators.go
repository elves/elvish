package complete

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"src.elv.sh/pkg/cli/lscolors"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/ui"
)

var pathSeparator = string(filepath.Separator)

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

func generateArgs(args []string, cfg Config) ([]RawItem, error) {
	switch args[0] {
	case "set", "tmp":
		for _, arg := range args[1:] {
			if arg == "=" {
				return nil, nil
			}
		}
		seed := args[len(args)-1]
		sigil, qname := eval.SplitSigil(seed)
		ns, _ := eval.SplitIncompleteQNameNs(qname)
		var items []RawItem
		cfg.PureEvaler.EachVariableInNs(ns, func(varname string) {
			items = append(items, noQuoteItem(sigil+parse.QuoteVariableName(ns+varname)))
		})
		return items, nil
	}

	items, err := cfg.ArgGenerator(args)
	return items, err
}

func generateExternalCommands(seed string, ev PureEvaler) ([]RawItem, error) {
	if fsutil.DontSearch(seed) {
		// Completing a local external command name.
		return generateFileNames(seed, true)
	}
	var items []RawItem
	ev.EachExternal(func(s string) { items = append(items, PlainItem(s)) })
	return items, nil
}

func generateCommands(seed string, ev PureEvaler) ([]RawItem, error) {
	if fsutil.DontSearch(seed) {
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

	sigil, qname := eval.SplitSigil(seed)
	ns, _ := eval.SplitIncompleteQNameNs(qname)
	if sigil == "" {
		// Generate functions, namespaces, and variable assignments.
		ev.EachVariableInNs(ns, func(varname string) {
			switch {
			case strings.HasSuffix(varname, eval.FnSuffix):
				addPlainItem(
					ns + varname[:len(varname)-len(eval.FnSuffix)])
			case strings.HasSuffix(varname, eval.NsSuffix):
				addPlainItem(ns + varname)
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

	files, err := os.ReadDir(dirToRead)
	if err != nil {
		return nil, fmt.Errorf("cannot list directory %s: %v", dirToRead, err)
	}

	lsColor := lscolors.GetColorist()

	// Make candidates out of elements that match the file component.
	for _, file := range files {
		name := file.Name()
		stat, err := file.Info()
		if err != nil {
			continue
		}
		// Show dot files iff file part of pattern starts with dot, and vice
		// versa.
		if dotfile(fileprefix) != dotfile(name) {
			continue
		}
		// Only accept searchable directories and executable files if
		// executableOnly is true.
		if onlyExecutable && (stat.Mode()&0111) == 0 {
			continue
		}

		// Full filename for source and getStyle.
		full := dir + name

		// Will be set to an empty space for non-directories
		suffix := " "

		if stat.IsDir() {
			full += pathSeparator
			suffix = ""
		} else if stat.Mode()&os.ModeSymlink != 0 {
			stat, err := os.Stat(full)
			if err == nil && stat.IsDir() { // symlink to directory
				full += pathSeparator
				suffix = ""
			}
		}

		items = append(items, ComplexItem{
			Stem:       full,
			CodeSuffix: suffix,
			Display:    ui.T(full, ui.StylingFromSGR(lsColor.GetStyle(full))),
		})
	}

	return items, nil
}

func generateIndices(v interface{}) []RawItem {
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
