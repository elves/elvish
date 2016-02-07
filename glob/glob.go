// Package glob implements globbing for elvish.
package glob

import (
	"io/ioutil"
	"os"
	"sort"
	"unicode/utf8"
)

type Pattern struct {
	Segments []Segment
}

type Segment struct {
	Type    SegmentType
	Literal string // Valid for Literal only.
}

type SegmentType int

const (
	Literal SegmentType = iota
	Star
	Question
	StarStar
	Slash
)

func Glob(p string) []string {
	return Parse(p).Glob()
}

func (p Pattern) Glob() []string {
	segs := p.Segments
	dir := ""
	if len(segs) > 0 && segs[0].Type == Slash {
		segs = segs[1:]
		dir = "/"
	}

	results := []string{}
	ch := make(chan string)
	go func() {
		glob(segs, dir, ch)
		close(ch)
	}()

	for s := range ch {
		results = append(results, s)
	}
	return results
}

func glob(segs []Segment, dir string, results chan<- string) {
	// Empty segment, resulting from a trailing slash. Generate the starting
	// directory.
	if len(segs) == 0 {
		results <- dir + "/"
		return
	}

	prefix := dir + "/"
	if dir == "" {
		prefix = ""
		dir = "."
	}

	i := -1
	nexti := func() {
		for i++; i < len(segs); i++ {
			if segs[i].Type == Slash || segs[i].Type == StarStar {
				break
			}
		}
	}
	nexti()

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		// XXX Silently drop the error
		return
	}

	// Enumerate the position of the first slash.
	for i < len(segs) {
		slash := segs[i].Type == Slash
		var first, rest []Segment
		if slash {
			// segs = x/y. Match dir with x, recurse on y.
			first, rest = segs[:i], segs[i+1:]
		} else {
			// segs = x**y. Match dir with x*, recurse on **y.
			first, rest = segs[:i+1], segs[i:]
		}

		for _, info := range infos {
			name := info.Name()
			if match(first, name) && info.IsDir() {
				glob(rest, prefix+name, results)
			}
		}

		if slash {
			// First slash cannot appear later than a slash in the pattern.
			return
		}
		nexti()
	}

	// If we reach here, it is possible to have no slashes at all.
	for _, info := range infos {
		name := info.Name()
		if match(segs, name) {
			results <- prefix + name
		}
	}
}

func filenames(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

// match matches a name against segments. It treats StarStar segments as they
// are Star segments. The segments may not contain Slash'es.
func match(segs []Segment, name string) bool {
segs:
	for len(segs) > 0 {
		// Find a chunk. A chunk is a run of Literal and Question, with an
		// optional leading Star.
		var i int
		for i = 1; i < len(segs); i++ {
			if segs[i].Type == Star || segs[i].Type == StarStar {
				break
			}
		}

		chunk := segs[:i]
		star := chunk[0].Type == Star || chunk[0].Type == StarStar
		if star {
			chunk = chunk[1:]
		}
		segs = segs[i:]

		// NOTE A quick path when len(segs) == 0 can be implemented: match
		// backwards.

		// Match at the current position. If this is the last chunk, we need to
		// make sure name is exhausted by the matching.
		ok, rest := matchChunk(chunk, name)
		if ok && (rest == "" || len(segs) > 0) {
			name = rest
			continue
		}

		if star {
			// NOTE An optimization is to make the upper bound not len(names),
			// but rather len(names) - LB(# bytes segs can match)
			for i := 1; i <= len(name); i++ {
				// Match after skipping i bytes.
				ok, rest := matchChunk(chunk, name[i:])
				if ok && (rest == "" || len(segs) > 0) {
					name = rest
					continue segs
				}
			}
		}
		return false
	}
	return name == ""
}

// matchChunk returns whether a chunk matches a prefix of name. If suceeded, it
// also returns the remaining part of name.
func matchChunk(chunk []Segment, name string) (bool, string) {
	for _, seg := range chunk {
		if name == "" {
			return false, ""
		}
		switch seg.Type {
		case Literal:
			n := len(seg.Literal)
			if len(name) < n || name[:n] != seg.Literal {
				return false, ""
			}
			name = name[n:]
		case Question:
			_, n := utf8.DecodeRuneInString(name)
			name = name[n:]
		}
	}
	return true, name
}
