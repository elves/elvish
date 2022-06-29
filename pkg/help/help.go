// Package help implements the `help` command.
package help

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"regexp"
	"sort"
	"strings"

	"src.elv.sh/pkg/edit/highlight"
	"src.elv.sh/pkg/ui"
)

type Text struct {
	display string
	search  string
}

var highlighter = highlight.NewHighlighter(highlight.Config{})

// Constants to communicate whether functions, variables, or both should be searched.
const (
	SearchFuncs = 1 << 0
	SearchVars  = 1 << 1
	SearchBoth  = SearchFuncs | SearchVars
)

// SearchWhich returns an int that encapsulates whether the documentation for functions, variables,
// or both should be searched. The quirk is that both as false is the same as both true.
func SearchWhich(f, v bool) int {
	switch {
	case (!f && !v) || (f && v):
		return SearchBoth
	case f:
		return SearchFuncs
	case v:
		return SearchVars
	}
	panic("impossible condition")
}

// Lookup returns the documentation for the named function or variable if it is known.
func Lookup(item string, which int) (Text, error) {
	if !strings.ContainsRune(item, ':') {
		item = "builtin:" + item
	}

	if which&SearchFuncs == SearchFuncs {
		if text, found := fnHelpDocs[item]; found {
			return text, nil
		}
	}
	if which&SearchVars == SearchVars {
		if text, found := varHelpDocs[item]; found {
			return text, nil
		}
	}
	return Text{}, errors.New("could not find documentation for " + item)
}

// EnumerateNamespace returns usage for each function and variable in a namespace; e.g. `builtin:`.
func EnumerateNamespace(termWidth int, namespace string, which int) ([]string, error) {
	var usages []string

	if which&SearchFuncs == SearchFuncs {
		for itemName, item := range fnHelpDocs {
			if strings.HasPrefix(itemName, namespace) {
				usages = append(usages, usageText(termWidth, true, itemName, item))
			}
		}
	}
	if which&SearchVars == SearchVars {
		for itemName, item := range varHelpDocs {
			if strings.HasPrefix(itemName, namespace) {
				usages = append(usages, usageText(termWidth, false, itemName, item))
			}
		}
	}
	if len(usages) != 0 {
		return usages, nil
	}
	return usages, errors.New("could not find documentation for namespace " + namespace)
}

func searchDocs(termWidth int, isFunction bool, documentation map[string]Text,
	searchTerms []string) []string {
	ignoreCase := true
	for _, searchTerm := range searchTerms {
		if searchTerm != strings.ToLower(searchTerm) {
			ignoreCase = false
			break
		}
	}

	var itemNames []string
	for itemName, item := range documentation {
		text := decompress(item.search)
		matched := true
		for _, searchTerm := range searchTerms {
			if ignoreCase {
				if !strings.Contains(strings.ToLower(text), searchTerm) {
					matched = false
					break
				}
			} else {
				if !strings.Contains(text, searchTerm) {
					matched = false
					break
				}
			}
		}
		if matched {
			itemNames = append(itemNames, itemName)
		}
	}

	sort.Strings(itemNames)
	var usages []string
	for _, itemName := range itemNames {
		usage := usageText(termWidth, isFunction, itemName, documentation[itemName])
		usages = append(usages, usage)
	}
	return usages
}

// Search returns the names of the functions and variables that contain the string(s).
func Search(termWidth int, which int, searchTerms []string) []string {
	var matchingFnDocs []string
	var matchingVarDocs []string

	if which&SearchFuncs == SearchFuncs {
		matchingFnDocs = searchDocs(termWidth, true, fnHelpDocs, searchTerms)
	}
	if which&SearchVars == SearchVars {
		matchingVarDocs = searchDocs(termWidth, false, varHelpDocs, searchTerms)
	}
	// We expand matchingVarDocs because it will will almost always be the much smaller slice.
	return append(matchingFnDocs, matchingVarDocs...)
}

// usageText returns a short, single line, "usage" for the item.
var condenseWhitespaceRe = regexp.MustCompile(`\s\s+`)

func usageText(termWidth int, isFunction bool, itemName string, item Text) string {
	// Drop the first line, and the empty second line since they don't provide useful information in
	// usage text.
	documentation := decompress(item.display)
	documentation = documentation[strings.IndexByte(documentation, '\n')+2:]

	if isFunction {
		// Ugh! This is too tightly coupled to the formatted help text for my taste but I don't see
		// a simpler solution. The "Usage:\n" prefix text protects against modifying
		// function/command documentation that does not start with the expected usage text.
		if strings.HasPrefix(documentation, "Usage:\n") {
			indentIdx := strings.IndexRune(documentation, ' ')
			if indentIdx != -1 {
				documentation = documentation[indentIdx+4:]
			}
		}
	} else {
		// This is documentation for a variable. Strip the "builtin:" prefix from the name for
		// consistency with the usage text for builtin functions.
		itemName = strings.TrimPrefix(itemName, "builtin:")
		highlighted, _ := highlighter.Get("$" + itemName)
		documentation = highlighted.String() + " " + documentation
	}

	documentation = strings.ReplaceAll(documentation, "\n", " ")
	builder := &strings.Builder{}
	var length int
	for _, segment := range ui.ParseSGREscapedText(documentation) {
		if length > termWidth {
			break
		}
		text := segment.Text
		text = strings.ReplaceAll(text, "\n", " ")
		text = condenseWhitespaceRe.ReplaceAllLiteralString(text, " ")
		styled := segment.Style.SGR() != ""
		if styled {
			length += len(text)
			if length > termWidth {
				break
			}
			segment.Text = text
			builder.WriteString(segment.String())
		} else {
			words := strings.Split(text, " ")
			nwords := len(words)
			for i, word := range strings.Split(text, " ") {
				if strings.HasPrefix(word, "Example:") || strings.HasPrefix(word, "Examples:") {
					length += 999999
					break // we've reached the end of a very short usage paragraph
				}
				length += len(word)
				if i > 0 && i < nwords {
					length += 1
				}
				if length > termWidth {
					break
				}
				if i > 0 && i < nwords {
					builder.WriteRune(' ')
				}
				builder.WriteString(word)
			}
		}
	}
	return builder.String() + "\n"
}

// DisplayText returns the documentation for the `item`.
func DisplayText(item Text) string {
	return decompress(item.display)
}

func decompress(s string) string {
	data, _ := base64.StdEncoding.DecodeString(s)
	rdata := bytes.NewReader(data)
	r, _ := gzip.NewReader(rdata)
	b, _ := ioutil.ReadAll(r)
	return string(b)
}
