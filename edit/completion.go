package edit

import (
	"fmt"
	"io/ioutil"

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

func findCandidates(p string, all []string) (cands []*candidate) {
	// Prefix match
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
	infos, e := ioutil.ReadDir(".")
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
	notPlainPrimary    = fmt.Errorf("not a plain PrimaryNode")
	notPlainCompound   = fmt.Errorf("not a plain CompoundNode")
	unknownContextType = fmt.Errorf("unknown context type")
)

func peekPrimary(fn *parse.PrimaryNode) (string, error) {
	if fn.Typ != parse.StringPrimary {
		return "", notPlainPrimary
	}
	return fn.Node.(*parse.StringNode).Text, nil
}

func peekIncompleteCompound(tn *parse.CompoundNode) (string, int, error) {
	text := ""
	for _, n := range tn.Nodes {
		if n.Right != nil {
			return "", 0, notPlainCompound
		}
		s, e := peekPrimary(n.Left)
		if e != nil {
			return "", 0, notPlainCompound
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
		return "", 0, unknownContextType
	}
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
	switch ctx.Typ {
	case parse.CommandContext:
		// BUG(xiaq): When completing, CommandContext is not supported
		ed.pushTip("command context not yet supported :(")
	case parse.NewArgContext, parse.ArgContext:
		// BUG(xiaq): When completing, [New]ArgContext is treated like RedirFilenameContext
		fallthrough
	case parse.RedirFilenameContext:
		// BUG(xiaq): When completing, only plain expressions are supported
		names, err := fileNames(".")
		if err != nil {
			ed.pushTip(err.Error())
			return nil
		}
		c.start = start
		c.end = ed.dot
		// BUG(xiaq) When completing, completion.typ is always ItemBare
		c.typ = parse.ItemBare
		c.candidates = findCandidates(compound, names)
		if len(c.candidates) > 0 {
			// TODO(xiaq): Support completions other than filename completion
			for _, c := range c.candidates {
				c.attr = defaultLsColor.determineAttr(c.text)
			}
			ed.completion = c
			ed.mode = modeCompletion
		} else {
			ed.pushTip(fmt.Sprintf("No completion for %s", compound))
		}
	}
	return nil
}
