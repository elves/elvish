// Package glob implements globbing for elvish.
package glob

import (
	"io/ioutil"
	"os"
	"unicode/utf8"
)

// Glob returns a list of file names satisfying the given pattern.
func Glob(p string, cb func(string) bool) bool {
	return Parse(p).Glob(cb)
}

// Glob returns a list of file names satisfying the Pattern.
func (p Pattern) Glob(cb func(string) bool) bool {
	segs := p.Segments
	dir := ""
	if len(segs) > 0 && IsSlash(segs[0]) {
		segs = segs[1:]
		dir = "/"
	}

	if p.DirOverride != "" {
		dir = p.DirOverride
	}

	return glob(segs, dir, cb)
}

func glob(segs []Segment, dir string, cb func(string) bool) bool {
	// Consume the non-wildcard prefix. This is required so that "." and "..",
	// which doesn't appear in the result of ReadDir, can appear as standalone
	// path components in the pattern.
	for len(segs) > 0 && IsLiteral(segs[0]) {
		seg0 := segs[0].(Literal).Data
		var path string
		switch dir {
		case "":
			path = seg0
		case "/":
			path = "/" + seg0
		default:
			path = dir + "/" + seg0
		}
		if len(segs) == 1 {
			// A lone literal. Generate it if the named file exists, and return.
			if _, err := os.Stat(path); err == nil {
				return cb(path)
			}
			return true
		} else if IsSlash(segs[1]) {
			// A lone literal followed by a slash. Change the directory if it
			// exists, otherwise return.
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				dir = path
			} else {
				return true
			}
			segs = segs[2:]
		} else {
			break
		}
	}

	// Empty segment, resulting from a trailing slash. Generate the starting
	// directory.
	if len(segs) == 0 {
		return cb(dir + "/")
	}

	var prefix string
	if dir == "" {
		prefix = ""
		dir = "."
	} else if dir == "/" {
		prefix = "/"
	} else {
		// dir never has a trailing slash unless it is /.
		prefix = dir + "/"
	}

	i := -1
	nexti := func() {
		for i++; i < len(segs); i++ {
			if IsSlash(segs[i]) || IsWild1(segs[i], StarStar) {
				break
			}
		}
	}
	nexti()

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		// XXX Silently drop the error
		return true
	}

	// Enumerate the position of the first slash.
	for i < len(segs) {
		slash := IsSlash(segs[i])
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
				if !glob(rest, prefix+name, cb) {
					return false
				}
			}
		}

		if slash {
			// First slash cannot appear later than a slash in the pattern.
			return true
		}
		nexti()
	}

	// If we reach here, it is possible to have no slashes at all.
	for _, info := range infos {
		name := info.Name()
		if match(segs, name) {
			if !cb(prefix + name) {
				return false
			}
		}
	}
	return true
}

// match matches a name against segments. It treats StarStar segments as they
// are Star segments. The segments may not contain Slash'es.
func match(segs []Segment, name string) bool {
	if len(segs) == 0 {
		return name == ""
	}
	// If the name start with "." and the first segment is a Wild, only match
	// when MatchHidden is true.
	if len(name) > 0 && name[0] == '.' && IsWild(segs[0]) {
		seg := segs[0].(Wild)
		if !seg.MatchHidden {
			return false
		}
	}
segs:
	for len(segs) > 0 {
		// Find a chunk. A chunk is a run of Literal and Question, with an
		// optional leading Star.
		var i int
		for i = 1; i < len(segs); i++ {
			if IsWild2(segs[i], Star, StarStar) {
				break
			}
		}

		chunk := segs[:i]
		startsWithStar := IsWild2(chunk[0], Star, StarStar)
		var startingStar Wild
		if startsWithStar {
			startingStar = chunk[0].(Wild)
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

		if startsWithStar {
			// NOTE An optimization is to make the upper bound not len(names),
			// but rather len(names) - LB(# bytes segs can match)
			for i, r := range name {
				j := i + len(string(r))
				// Match name[:j] with the starting *, and the rest with chunk.
				if !startingStar.Match(r) {
					break
				}
				ok, rest := matchChunk(chunk, name[j:])
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
		switch seg := seg.(type) {
		case Literal:
			n := len(seg.Data)
			if len(name) < n || name[:n] != seg.Data {
				return false, ""
			}
			name = name[n:]
		case Wild:
			if seg.Type == Question {
				r, n := utf8.DecodeRuneInString(name)
				if !seg.Match(r) {
					return false, ""
				}
				name = name[n:]
			} else {
				panic("chunk has non-question wild segment")
			}
		default:
			panic("chunk has non-literal non-wild segment")
		}
	}
	return true, name
}
