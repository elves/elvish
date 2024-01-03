package os

import (
	"io/fs"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

// Conversion between an Elvish list representation of special modes and Go's
// bit flag representation. These are used from different places, but since they
// are symmetrical, keeping them in the same file makes it easier to ensure they
// are consistent.
//
// A special mode is one of the bits in [fs.FileMode] that is not part of
// [fs.ModeType] or [fs.ModePerm]. We omit [fs.ModeAppend], [fs.ModeExclusive]
// and [fs.ModeTemporary] since they are only used on Plan 9, which Elvish
// doesn't support (yet) so we can't test them.
//
// TODO: Use a set as the Elvish representation when Elvish has lists.

func specialModesFromIterable(v any) (fs.FileMode, error) {
	var mode fs.FileMode
	var errElem error
	errIterate := vals.Iterate(v, func(elem any) bool {
		switch elem {
		case "setuid":
			mode |= fs.ModeSetuid
		case "setgid":
			mode |= fs.ModeSetgid
		case "sticky":
			mode |= fs.ModeSticky
		default:
			errElem = errs.BadValue{What: "special mode",
				Valid: "setuid, setgid or sticky", Actual: vals.ToString(elem)}
			return false
		}
		return true
	})
	if errIterate != nil {
		return 0, errIterate
	}
	if errElem != nil {
		return 0, errElem
	}
	return mode, nil
}

func specialModesToList(mode fs.FileMode) vals.List {
	l := vals.EmptyList
	if mode&fs.ModeSetuid != 0 {
		l = l.Conj("setuid")
	}
	if mode&fs.ModeSetgid != 0 {
		l = l.Conj("setgid")
	}
	if mode&fs.ModeSticky != 0 {
		l = l.Conj("sticky")
	}
	return l
}
