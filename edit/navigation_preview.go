package edit

import (
	"errors"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/util"
)

// PreviewBytes is the maximum number of bytes to preview a file.
const PreviewBytes = 64 * 1024

// Errors displayed in the preview area.
var (
	ErrNotRegular   = errors.New("no preview for non-regular file")
	ErrNotValidUTF8 = errors.New("no preview for non-utf8 file")
)

type navFilePreview struct {
	lines     []ui.Styled
	fullWidth int
	beginLine int
}

func newNavFilePreview(lines []string) *navFilePreview {
	width := 0
	convertedLines := make([]ui.Styled, len(lines))
	for i, line := range lines {
		// BUG: Handle tabstops correctly
		convertedLine := strings.Replace(line, "\t", "    ", -1)
		convertedLines[i] = ui.Unstyled(convertedLine)
		width = max(width, util.Wcswidth(convertedLine))
	}
	return &navFilePreview{convertedLines, width, 0}
}

func (fp *navFilePreview) FullWidth(h int) int {
	width := fp.fullWidth
	if h < len(fp.lines) {
		return width + 1
	}
	return width
}

func (fp *navFilePreview) List(h int) ui.Renderer {
	if len(fp.lines) <= h {
		logger.Printf("Height %d fit all lines", h)
		return listingRenderer{fp.lines}
	}
	shown := fp.lines[fp.beginLine:]
	if len(shown) > h {
		shown = shown[:h]
	}
	logger.Printf("Showing lines %d to %d", fp.beginLine, fp.beginLine+len(shown))
	return listingWithScrollBarRenderer{
		listingRenderer{shown}, len(fp.lines),
		fp.beginLine, fp.beginLine + len(shown), h}
}

func makeNavFilePreview(fname string) navPreview {
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

	return newNavFilePreview(strings.Split(content, "\n"))
}
