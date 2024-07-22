package complete

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"src.elv.sh/pkg/cli/lscolors"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/np"
	"src.elv.sh/pkg/ui"
)

const pathSeparator = string(filepath.Separator)

var eachExternal = fsutil.EachExternal

// GenerateFileNames returns filename candidates that are suitable for completing
// the last argument. It can be used in Config.ArgGenerator.
func GenerateFileNames(args []string) ([]RawItem, error) {
	if len(args) == 0 {
		return nil, nil
	}
	return generateFileNames(args[len(args)-1], nil)
}

// GenerateFileNames returns directory name candidates that are suitable for
// completing the last argument. It can be used in Config.ArgGenerator.
func GenerateDirNames(args []string) ([]RawItem, error) {
	if len(args) == 0 {
		return nil, nil
	}
	return generateFileNames(args[len(args)-1], fs.FileInfo.IsDir)
}

// GenerateForSudo generates candidates for sudo.
func GenerateForSudo(args []string, ev *eval.Evaler, cfg Config) ([]RawItem, error) {
	switch {
	case len(args) < 2:
		return nil, errNoCompletion
	case len(args) == 2:
		// Complete external commands.
		return generateExternalCommands(args[1])
	default:
		return cfg.ArgGenerator(args[1:])
	}
}

// Internal generators, used from completers.

func generateArgs(args []string, ev *eval.Evaler, p np.Path, cfg Config) ([]RawItem, error) {
	switch args[0] {
	case "set", "tmp":
		for i := 1; i < len(args); i++ {
			if args[i] == "=" {
				if i == len(args)-1 {
					// Completing the "=" itself; don't offer any candidates.
					return nil, nil
				} else {
					// Completing an argument after "="; fall back to the
					// default arg generator.
					return cfg.ArgGenerator(args)
				}
			}
		}
		seed := args[len(args)-1]
		sigil, qname := eval.SplitSigil(seed)
		ns, _ := eval.SplitIncompleteQNameNs(qname)
		var items []RawItem
		eachVariableInNs(ev, p, ns, func(varname string) {
			items = append(items, noQuoteItem(sigil+parse.QuoteVariableName(ns+varname)))
		})
		return items, nil
	case "del":
		// This partially duplicates eachVariableInNs with ns = "", but we don't
		// offer builtin variables.
		var items []RawItem
		addItem := func(varname string) {
			items = append(items, noQuoteItem(parse.QuoteVariableName(varname)))
		}
		ev.Global().IterateKeysString(addItem)
		eachDefinedVariable(p[len(p)-1], p[0].Range().From, addItem)
		return items, nil
	}

	return cfg.ArgGenerator(args)
}

func generateExternalCommands(seed string) ([]RawItem, error) {
	if fsutil.DontSearch(seed) {
		// Completing a local external command name.
		return generateFileNames(seed, executableOrDir)
	}
	var items []RawItem
	eachExternal(func(s string) { items = append(items, PlainItem(s)) })
	return items, nil
}

func generateCommands(seed string, ev *eval.Evaler, p np.Path) ([]RawItem, error) {
	if fsutil.DontSearch(seed) {
		// Completing a local external command name.
		return generateFileNames(seed, executableOrDir)
	}

	var cands []RawItem
	addPlainItem := func(s string) { cands = append(cands, PlainItem(s)) }

	if strings.HasPrefix(seed, "e:") {
		// Generate all external commands with the e: prefix, and be done.
		eachExternal(func(command string) {
			addPlainItem("e:" + command)
		})
		return cands, nil
	}

	// Generate all special forms.
	for name := range eval.IsBuiltinSpecial {
		addPlainItem(name)
	}
	// Generate all external commands (without the e: prefix).
	eachExternal(addPlainItem)

	sigil, qname := eval.SplitSigil(seed)
	ns, _ := eval.SplitIncompleteQNameNs(qname)
	if sigil == "" {
		// Generate functions, namespaces, and variable assignments.
		eachVariableInNs(ev, p, ns, func(varname string) {
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

func generateFileNames(seed string, statPred func(fs.FileInfo) bool) ([]RawItem, error) {
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
		// Apply statPred if given.
		if statPred != nil && !statPred(stat) {
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

func executableOrDir(stat fs.FileInfo) bool {
	return fsutil.IsExecutable(stat) || stat.IsDir()
}

func generateIndices(v any) []RawItem {
	var items []RawItem
	vals.IterateKeys(v, func(k any) bool {
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
