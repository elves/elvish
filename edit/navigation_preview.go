package edit

import (
	"errors"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/edit/ui"
)

// PreviewBytes is the maximum number of bytes to preview a file.
const PreviewBytes = 64 * 1024

// Errors displayed in the preview area.
var (
	ErrNotRegular   = errors.New("no preview for non-regular file")
	ErrNotValidUTF8 = errors.New("no preview for non-utf8 file")
)

func newFilePreviewNavColumn(fname string) navPreview {
	file, err := os.Open(fname)
	if err != nil {
		return newErrNavColumn(err)
	}

	info, err := file.Stat()
	if err != nil {
		return newErrNavColumn(err)
	}
	if (info.Mode() & (os.ModeDevice | os.ModeNamedPipe | os.ModeSocket | os.ModeCharDevice)) != 0 {
		return newErrNavColumn(ErrNotRegular)
	}

	// BUG: when the file is bigger than the buffer, the scrollbar is wrong.
	var buf [PreviewBytes]byte
	nr, err := file.Read(buf[:])
	if err != nil {
		return newErrNavColumn(err)
	}

	content := string(buf[:nr])
	if !utf8.ValidString(content) {
		return newErrNavColumn(ErrNotValidUTF8)
	}

	lines := strings.Split(content, "\n")
	styleds := make([]ui.Styled, len(lines))
	for i, line := range lines {
		styleds[i] = ui.Styled{strings.Replace(line, "\t", "    ", -1), ui.Styles{}}
	}
	return newNavColumn(styleds, func(int) bool { return false })
}
