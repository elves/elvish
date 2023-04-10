// Package glob implements globbing for elvish.
package glob

import (
	"os"
	"runtime"
	"unicode/utf8"
)

// TODO: On Windows, preserve the original path separator (/ or \) specified in
// the glob pattern.

// PathInfo keeps a path resulting from glob expansion and its FileInfo. The
// FileInfo is useful for efficiently determining if a given pathname satisfies
// a particular constraint without doing an extra stat.
type PathInfo struct {
	// The generated path, consistent with the original glob pattern. It cannot
	// be replaced by Info.Name(), which is just the final path component.
	Path string
	Info os.FileInfo
}

// Glob returns a list of file names satisfying the given pattern.
func Glob(p string, cb func(PathInfo) bool) bool {
	return Parse(p).Glob(cb)
}

// Glob returns a list of file names satisfying the Pattern.
func (p Pattern) Glob(cb func(PathInfo) bool) bool {
	segs := p.Segments
	dir := ""

	// TODO(xiaq): This is a hack solely for supporting globs that start with
	// ~ (tilde) in the eval package.
	if p.DirOverride != "" {
		dir = p.DirOverride
	}

	if len(segs) > 0 && IsSlash(segs[0]) {
		segs = segs[1:]
		dir += "/"
	} else if runtime.GOOS == "windows" && len(segs) > 1 && IsLiteral(segs[0]) && IsSlash(segs[1]) {
		// TODO: Handle Windows UNC paths.
		elem := segs[0].(Literal).Data
		if isDrive(elem) {
			segs = segs[2:]
			dir = elem + "/"
		}
	}

	return glob(segs, dir, cb)
}

// isLetter returns true if the byte is an ASCII letter.
func isLetter(chr byte) bool {
	return ('a' <= chr && chr <= 'z') || ('A' <= chr && chr <= 'Z')
}

// isDrive returns true if the string looks like a Windows drive letter path prefix.
func isDrive(s string) bool {
	return len(s) == 2 && s[1] == ':' && isLetter(s[0])
}

// glob finds all filenames matching the given Segments in the given dir, and
// calls the callback on all of them. If the callback returns false, globbing is
// interrupted, and glob returns false. Otherwise it returns true. Files that
// can't be lstat'ed and directories that can't be read are ignored silently.
func glob(segs []Segment, dir string, cb func(PathInfo) bool) bool {
	// Consume non-wildcard path elements simply by following the path. This may
	// seem like an optimization, but is actually required for "." and ".." to
	// be used as path elements, as they do not appear in the result of ReadDir.
	// It is also required for handling directory components that are actually
	// symbolic links to directories.
	for len(segs) > 1 && IsLiteral(segs[0]) && IsSlash(segs[1]) {
		elem := segs[0].(Literal).Data
		segs = segs[2:]
		dir += elem + "/"
		// This will correctly resolve symbolic links when they appear literally
		// (e.g. in "link-to-dir/*") despite the use of Lstat, since a trailing
		// slash always causes symbolic links to be resolved
		// (https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap04.html#tag_04_13).
		if info, err := os.Lstat(dir); err != nil || !info.IsDir() {
			return true
		}
	}

	if len(segs) == 0 {
		if info, err := os.Lstat(dir); err == nil {
			return cb(PathInfo{dir, info})
		}
		return true
	} else if len(segs) == 1 && IsLiteral(segs[0]) {
		path := dir + segs[0].(Literal).Data
		if info, err := os.Lstat(path); err == nil {
			return cb(PathInfo{path, info})
		}
		return true
	}

	infos, err := readDir(dir)
	if err != nil {
		// Ignore directories that can't be read.
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
			fullname := dir + name
			info, err := os.Lstat(fullname)
			if err != nil {
				// Either the file was removed between ReadDir and Lstat, or the
				// OS has some special rule that prevents it from being lstat'ed
				// (see b.elv.sh/1674 for a known case on macOS; SELinux and
				// FreeBSD's MAC might be able to do the same). In either case,
				// ignore the file.
				continue
			}
			if !cb(PathInfo{fullname, info}) {
				return false
			}
		}
	}
	return true
}

// readDir is just like os.ReadDir except that it treats an argument of "" as ".".
func readDir(dir string) ([]os.DirEntry, error) {
	if dir == "" {
		dir = "."
	}
	return os.ReadDir(dir)
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

		// TODO: Implement a quick path when len(segs) == 0 by matching
		// backwards.

		// Match at the current position. If this is the last chunk, we need to
		// make sure name is exhausted by the matching.
		ok, rest := matchFixedLength(chunk, name)
		if ok && (rest == "" || len(segs) > 0) {
			name = rest
			continue
		}

		if startsWithStar {
			// TODO: Optimize by stopping at len(name) - LB(# bytes segs can
			// match) rather than len(names)
			for i := 0; i < len(name); {
				r, rsize := utf8.DecodeRuneInString(name[i:])
				j := i + rsize
				// Match name[:j] with the starting *, and the rest with chunk.
				if !startingStar.Match(r) {
					break
				}
				ok, rest := matchFixedLength(chunk, name[j:])
				if ok && (rest == "" || len(segs) > 0) {
					name = rest
					continue segs
				}
				i = j
			}
		}
		return false
	}
	return name == ""
}

// matchFixedLength returns whether a run of fixed-length segments (Literal and
// Question) matches a prefix of name. It returns whether the match is
// successful and if it is, the remaining part of name.
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
