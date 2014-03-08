package edit

import (
	"fmt"
	"io/ioutil"

	"github.com/xiaq/elvish/parse"
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

func startCompletion(ed *Editor, k Key) *leReturn {
	c := &completion{}
	ctx, err := parse.Complete("<completion>", ed.line[:ed.dot])
	if err != nil {
		ed.pushTip("parser error")
		return nil
	}
	pctx := ctx.EvalPlain()
	if pctx == nil {
		ed.pushTip("context not plain")
		return nil
	}
	switch pctx.Typ {
	case parse.CommandContext:
		// BUG(xiaq): When completing, CommandContext is not supported
		ed.pushTip("command context not yet supported :(")
	case parse.ArgContext:
		// BUG(xiaq): When completing, ArgContext is treated like RedirFilenameContext
		fallthrough
	case parse.RedirFilenameContext:
		// BUG(xiaq): When completing, only the case of ctx.ThisFactor.Typ == StringFactor is supported
		if pctx.ThisFactor.Typ != parse.StringFactor {
			ed.pushTip("only StringFactor is supported :(")
			return nil
		}
		pattern := pctx.PrevFactors + pctx.ThisFactor.Node.(*parse.StringNode).Text
		names, err := fileNames(".")
		if err != nil {
			ed.pushTip(err.Error())
			return nil
		}
		c.start = int(ctx.PrevFactors.Pos)
		c.end = ed.dot
		// BUG(xiaq) When completing, completion.typ is always ItemBare
		c.typ = parse.ItemBare
		c.candidates = findCandidates(pattern, names)
		if len(c.candidates) > 0 {
			// XXX assumes filename candidate
			for _, c := range c.candidates {
				c.attr = defaultLsColor.determineAttr(c.text)
			}
			ed.completion = c
			ed.mode = modeCompletion
		} else {
			ed.pushTip(fmt.Sprintf("No completion for %s", pattern))
		}
	}
	return nil
}
