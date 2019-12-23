// Package glob implements globbing for elvish.
package glob

import (
	"io/ioutil"
	"os"
	"runtime"
	"unicode/utf8"
)

// TODO: Use native path separators instead of always using /.

// Glob returns a list of file names satisfying the given pattern.
func Glob(p string, cb func(string) bool) bool {
	return Parse(p).Glob(cb)
}

// Glob returns a list of file names satisfying the Pattern.
func (p Pattern) Glob(cb func(string) bool) bool {
	segs := p.Segments
	dir := ""

	// XXX: This is a hack solely for supporting globs that start with ~ in the
	// eval package.
	if p.DirOverride != "" {
		dir = p.DirOverride
	}

	if len(segs) > 0 && IsSlash(segs[0]) {
		segs = segs[1:]
		dir += "/"
	} else if runtime.GOOS == "windows" && len(segs) > 1 && IsLiteral(segs[0]) && IsSlash(segs[1]) {
		// TODO: Handle UNC.
		elem := segs[0].(Literal).Data
		if isDrive(elem) {
			segs = segs[2:]
			dir = elem + "/"
		}
	}

	return glob(segs, dir, cb)
}

func isDrive(s string) bool {
	return len(s) == 2 && s[1] == ':' &&
		(('a' <= s[0] && s[1] <= 'z') || ('A' <= s[0] && s[0] <= 'Z'))
}

// glob finds all filenames matching the given Segments in the given dir, and
// calls the callback on all of them. If the callback returns false, globbing is
// interrupted, and glob returns false. Otherwise it returns true.
func glob(segs []Segment, dir string, cb func(string) bool) bool {
	// Consume non-wildcard path elements simply by following the path. This may
	// seem like an optimization, but is actually required for "." and ".." to
	// be used as path elements, as they do not appear in the result of ReadDir.
	for len(segs) > 1 && IsLiteral(segs[0]) && IsSlash(segs[1]) {
		elem := segs[0].(Literal).Data
		segs = segs[2:]
		dir += elem + "/"
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			return true
		}
	}

	if len(segs) == 0 {
		return cb(dir)
	} else if len(segs) == 1 && IsLiteral(segs[0]) {
		path := dir + segs[0].(Literal).Data
		if _, err := os.Stat(path); err == nil {
			return cb(path)
		}
		return true
	}

	infos, err := readDir(dir)
	if err != nil {
		// XXX Silently drop the error
		return true
	}

	i := -1
	// nexti moves i to the next index in segs that is either / or ** (in other
	// words, something that matches /).
	nexti := func() {
		for i++; i < len(segs); i++ {
			if IsSlash(segs[i]) || IsWild1(segs[i], StarStar) {
				break
			}
		}
	}
	nexti()

	// Enumerate the position of the first slash. In the presence of multiple
	// **'s in the pattern, the first slash may be in any of those.
	//
	// For instance, in x**y**z, the first slash may be in the first ** or the
	// second:
	// 1) If it is in the first, then pattern is equivalent to x*/**y**z. We
	//    match directories with x* and recurse in each subdirectory with the
	//    pattern **y**z.
	// 2) If it is the in the second, we know that since the first ** can no
	//    longer contain any slashes, we treat it as * (this is done in
	//    matchElement). The pattern is now equivalent to x*y*/**z. We match
	//    directories with x*y* and recurse in each subdirectory with the
	//    pattern **z.
	//
	// The rules are:
	// 1) For each **, we treat it as */** and all previous ones as *. We match
	//    subdirectories with the part before /, and recurse in subdirectories
	//    with the pattern after /.
	// 2) If a literal / is encountered, we return after recursing in the
	//    subdirectories.
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
			if matchElement(first, name) && info.IsDir() {
				if !glob(rest, dir+name+"/", cb) {
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

	// If we reach here, it is possible to have no slashes at all. Simply match
	// the entire pattern with all files.
	for _, info := range infos {
		name := info.Name()
		if matchElement(segs, name) {
			if !cb(dir + name) {
				return false
			}
		}
	}
	return true
}

// readDir is just like ioutil.ReadDir except that it treats an argument of ""
// as ".".
func readDir(dir string) ([]os.FileInfo, error) {
	if dir == "" {
		dir = "."
	}
	return ioutil.ReadDir(dir)
}

// matchElement matches a path element against segments, which may not contain
// any Slash segments. It treats StarStar segments as they are Star segments.
func matchElement(segs []Segment, name string) bool {
	if len(segs) == 0 {
		return name == ""
	}
	// If the name start with "." and the first segment is a Wild, only match
	// when MatchHidden is true.
	if len(name) > 0 && name[0] == '.' && IsWild(segs[0]) && !segs[0].(Wild).MatchHidden {
		return false
	}
segs:
	for len(segs) > 0 {
		// Find a chunk. A chunk is an optional Star followed by a run of
		// fixed-length segments (Literal and Question).
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
		ok, rest := matchFixedLength(chunk, name)
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
				ok, rest := matchFixedLength(chunk, name[j:])
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

// matchFixedLength returns whether a run of fixed-length segments (Literal and
// Question) matches a prefix of name. It returns whether the match is
// successful and if if it is, the remaining part of name.
func matchFixedLength(segs []Segment, name string) (bool, string) {
	for _, seg := range segs {
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
				panic("matchFixedLength given non-question wild segment")
			}
		default:
			panic("matchFixedLength given non-literal non-wild segment")
		}
	}
	return true, name
}
