package edit

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type tokenPart struct {
	text      string
	completed bool
}

type candidate struct {
	text  string
	parts []tokenPart
	attr  string // Attribute used for preview
}

func newCandidate() *candidate {
	return &candidate{}
}

func (c *candidate) push(tp tokenPart) {
	c.text += tp.text
	c.parts = append(c.parts, tp)
}

type completion struct {
	start, end int // The text to complete is Editor.line[start:end]
	typ        parse.ItemType
	candidates []*candidate
	current    int
}

func (c *completion) prev(cycle bool) {
	c.current--
	if c.current == -1 {
		if cycle {
			c.current = len(c.candidates) - 1
		} else {
			c.current++
		}
	}
}

func (c *completion) next(cycle bool) {
	c.current++
	if c.current == len(c.candidates) {
		if cycle {
			c.current = 0
		} else {
			c.current--
		}
	}
}

// Find candidates by matching against a prefix.
func prefixMatchCandidates(p string, all []string) (cands []*candidate) {
	for _, s := range all {
		if len(s) >= len(p) && s[:len(p)] == p {
			cand := newCandidate()
			cand.push(tokenPart{p, false})
			cand.push(tokenPart{s[len(p):], true})
			cands = append(cands, cand)
		}
	}
	return
}

func fileNames(dir string) (names []string, err error) {
	infos, e := ioutil.ReadDir(dir)
	if e != nil {
		err = e
		return
	}
	for _, info := range infos {
		names = append(names, info.Name())
	}
	return
}

var (
	errNotPlainPrimary    = fmt.Errorf("not a plain PrimaryNode")
	errNotPlainCompound   = fmt.Errorf("not a plain CompoundNode")
	errUnknownContextType = fmt.Errorf("unknown context type")
)

func peekPrimary(fn *parse.Primary) (string, error) {
	if fn.Typ != parse.StringPrimary {
		return "", errNotPlainPrimary
	}
	return fn.Node.(*parse.String).Text, nil
}

func peekIncompleteCompound(tn *parse.Compound) (string, int, error) {
	text := ""
	for _, n := range tn.Nodes {
		if n.Right != nil {
			return "", 0, errNotPlainCompound
		}
		s, e := peekPrimary(n.Left)
		if e != nil {
			return "", 0, errNotPlainCompound
		}
		text += s
	}
	return text, int(tn.Pos), nil
}

func peekCurrentCompound(ctx *parse.Context, dot int) (string, int, error) {
	if ctx.Form == nil || ctx.Typ == parse.NewArgContext {
		return "", dot, nil
	}

	switch ctx.Typ {
	case parse.CommandContext:
		return peekIncompleteCompound(ctx.Form.Command)
	case parse.ArgContext:
		compounds := ctx.Form.Args.Nodes
		lastCompound := compounds[len(compounds)-1]
		return peekIncompleteCompound(lastCompound)
	case parse.RedirFilenameContext:
		redirs := ctx.Form.Redirs
		lastRedir := redirs[len(redirs)-1]
		fnRedir, ok := lastRedir.(*parse.FilenameRedir)
		if !ok {
			return "", 0, fmt.Errorf("last redir is not FilenameRedir")
		}
		return peekIncompleteCompound(fnRedir.Filename)
	default:
		return "", 0, errUnknownContextType
	}
}

var builtins []string

func init() {
	builtins = append(builtins, eval.BuiltinFnNames...)
	builtins = append(builtins, eval.BuiltinSpecialNames...)
}

func startCompletion(ed *Editor, k Key) *leReturn {
	c := &completion{}
	ctx, err := parse.Complete("<completion>", ed.line[:ed.dot])
	if err != nil {
		ed.pushTip("parser error")
		return nil
	}
	compound, start, err := peekCurrentCompound(ctx, ed.dot)
	if err != nil {
		ed.pushTip("cannot complete :(")
		return nil
	}

	var findAll func() ([]string, error)
	var makeCandidates func([]string) []*candidate

	switch ctx.Typ {
	case parse.CommandContext:
		// BUG(xiaq): When completing commands, only builtins are searched
		findAll = func() ([]string, error) {
			return builtins, nil
		}
		makeCandidates = func(all []string) (cands []*candidate) {
			return prefixMatchCandidates(compound, all)
		}
	case parse.NewArgContext, parse.ArgContext:
		// BUG(xiaq): When completing, [New]ArgContext is treated like RedirFilenameContext
		fallthrough
	case parse.RedirFilenameContext:
		// BUG(xiaq): File name completion does not deal with meta characters.
		// Assume that compound is an incomplete filename
		dir, file := path.Split(compound)
		findAll = func() ([]string, error) {
			return fileNames(dir)
		}
		makeCandidates = func(all []string) (cands []*candidate) {
			// Make candidates out of elements that match the file component.
			for _, s := range all {
				if strings.HasPrefix(s, file) {
					cand := newCandidate()
					cand.push(tokenPart{compound, false})
					cand.push(tokenPart{s[len(file):], true})
					cand.attr = defaultLsColor.determineAttr(cand.text)
					cands = append(cands, cand)
				}
			}
			return
		}
	}
	// BUG(xiaq): When completing, only plain expressions are supported
	all, err := findAll()
	if err != nil {
		ed.pushTip(err.Error())
		return nil
	}
	c.start = start
	c.end = ed.dot
	// BUG(xiaq) When completing, completion.typ is always ItemBare
	c.typ = parse.ItemBare
	c.candidates = makeCandidates(all)
	if len(c.candidates) > 0 {
		ed.completion = c
		ed.mode = modeCompletion
	} else {
		ed.pushTip(fmt.Sprintf("No completion for %s", compound))
	}
	return nil
}
