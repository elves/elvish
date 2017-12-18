package edit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/elves/elvish/edit/nodeutil"

	"github.com/elves/elvish/edit/lscolors"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type argComplContext struct {
	seed       string
	quoting    parse.PrimaryType
	words      []string
	begin, end int
}

func (*argComplContext) name() string { return "argument" }

func findArgComplContext(n parse.Node, ev *eval.Evaler) complContext {
	if sep, ok := n.(*parse.Sep); ok {
		if form, ok := sep.Parent().(*parse.Form); ok && form.Head != nil {
			return &argComplContext{
				"", quotingForEmptySeed,
				evalFormPure(form, "", n.End(), ev),
				n.End(), n.End()}
		}
	}
	if primary, ok := n.(*parse.Primary); ok {
		if compound, seed := primaryInSimpleCompound(primary, ev); compound != nil {
			if form, ok := compound.Parent().(*parse.Form); ok {
				if form.Head != nil && form.Head != compound {
					return &argComplContext{
						seed, primary.Type,
						evalFormPure(form, seed, compound.Begin(), ev),
						compound.Begin(), compound.End()}
				}
			}
		}
	}
	return nil
}

func evalFormPure(form *parse.Form, seed string, seedBegin int, ev *eval.Evaler) []string {
	// Find out head of the form and preceding arguments.
	// If form.Head is not a simple compound, head will be "", just what we want.
	head, _ := nodeutil.PurelyEvalPartialCompound(form.Head, nil, ev)
	words := []string{head}
	for _, compound := range form.Args {
		if compound.Begin() >= seedBegin {
			break
		}
		if arg, err := nodeutil.PurelyEvalCompound(compound, ev); err == nil {
			// XXX Arguments that are not simple compounds are simply ignored.
			words = append(words, arg)
		}
	}

	words = append(words, seed)
	return words
}

// To complete an argument, delegate the actual completion work to a suitable
// complContext.
func (ctx *argComplContext) complete(ev *eval.Evaler, matcher eval.CallableValue) (*complSpec, error) {
	rawCands := make(chan rawCandidate)
	collectErr := make(chan error)
	go func() {
		var err error
		defer func() {
			close(rawCands)
			collectErr <- err
		}()

		err = completeArg(ctx.words, ev, rawCands)
	}()

	cands, err := ev.Editor.(*Editor).filterAndCookCandidates(
		ev, matcher, ctx.seed, rawCands, ctx.quoting)
	if ce := <-collectErr; ce != nil {
		return nil, ce
	}
	if err != nil {
		return nil, err
	}

	return &complSpec{ctx.begin, ctx.end, cands}, nil
}

// TODO: getStyle does redundant stats.
func complFilenameInner(head string, executableOnly bool, rawCands chan<- rawCandidate) error {
	dir, fileprefix := filepath.Split(head)
	dirToRead := dir
	if dirToRead == "" {
		dirToRead = "."
	}

	infos, err := ioutil.ReadDir(dirToRead)
	if err != nil {
		return fmt.Errorf("cannot list directory %s: %v", dirToRead, err)
	}

	lsColor := lscolors.GetColorist()
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
		if executableOnly && !(info.IsDir() || (info.Mode()&0111) != 0) {
			continue
		}

		// Full filename for source and getStyle.
		full := dir + name

		suffix := " "
		if info.IsDir() {
			suffix = string(filepath.Separator)
		} else if info.Mode()&os.ModeSymlink != 0 {
			stat, err := os.Stat(full)
			if err == nil && stat.IsDir() {
				// Symlink to directory.
				suffix = string(filepath.Separator)
			}
		}

		rawCands <- &complexCandidate{
			stem: full, codeSuffix: suffix,
			style: ui.StylesFromString(lsColor.GetStyle(full)),
		}
	}

	return nil
}

func dotfile(fname string) bool {
	return strings.HasPrefix(fname, ".")
}
