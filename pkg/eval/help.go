package eval

import (
	"sort"
	"strings"

	"src.elv.sh/pkg/help"
	"src.elv.sh/pkg/sys"
)

func init() {
	addBuiltinFns(map[string]any{
		"help": helpCmd,
	})
}

//elvdoc:fn help
//
// ```elvish
// help &fn=$false &var=$false &search=$false $search-term...
// ```
//
// Output help text for the function or variable matching `$search-term`. If no search terms are
// provided the documentation for the `builtin:help` command is output.
//
// If `&search` is false, and there is a single search term, the documentation for the item is
// output. If `$search-term` matches both a function and variable (unlikely but possible) only the
// function definition is output unless the `&var` option is used. The `builtin:` namespace is
// searched implicitly if the search term does not have a namespace prefix. If the lookup fails a
// search is performed.
//
// If there is a single search term and it looks like a namespace, such as `math:`, a usage line for
// each item in the namespace is output. Subject to the use of the `&fn` and `&var` options to limit
// which items are selected for output.
//
// Use `&search` to search the documentation for each `$search-term`. This defaults to `$false` if
// there is zero or one search term and `$true` if there is more than one search term. The search is
// case-insensitive unless at least one search term includes at least one uppercase character. The
// search term is a string that is matched literally without regard to word boundaries. Each search
// term is searched in the documentation independent of the preceding term; i.e., the search does
// not require the terms to be adjacent in the documentation but the documentation must contain all
// search terms.
//
// Use `&fn` to restrict the lookup, or search, to functions only.
//
// Use `&var` to restrict the lookup, or search, to variables only.
//
// If both `&fn` and `&var` are false that is equivalent to setting both true.

type helpOpts struct{ Search, Var, Fn bool }

func (o *helpOpts) SetDefaultOptions() {}

func helpCmd(fm *Frame, opts helpOpts, searchTerms ...string) error {
	out := fm.ByteOutput()
	which := help.SearchWhich(opts.Fn, opts.Var)

	if len(searchTerms) == 0 {
		item, err := help.Lookup("builtin:help", help.SearchFuncs)
		if err == nil {
			_, err = out.WriteString(help.DisplayText(item))
		}
		return err
	} else if len(searchTerms) > 1 {
		opts.Search = true
	}

	in := fm.InputFile()
	var termWidth int = 80
	if sys.IsATTY(in) {
		if _, width := sys.WinSize(in); width > 1 {
			termWidth = width
		}
	}

	if !opts.Search {
		// Output the documentation for a specific namespace, function or variable.
		searchTerm := searchTerms[0]
		if searchTerm[len(searchTerm)-1] == ':' { // enumerate items in a namespace
			usages, err := help.EnumerateNamespace(termWidth, searchTerm, which)
			if err == nil {
				sort.Strings(usages)
				_, err = out.WriteString(strings.Join(usages, ""))
			}
			return err
		} else { // find documentation for a specific function and/or variable
			item, err := help.Lookup(searchTerm, which)
			if err == nil {
				_, err = out.WriteString(help.DisplayText(item))
				return err
			}
		}
	}

	// Search the documentation for strings matching the search terms and output the usage for the
	// matching items.
	matches := help.Search(termWidth, which, searchTerms)
	for _, match := range matches {
		if _, err := out.WriteString(match); err != nil {
			return err
		}
	}

	return nil
}
