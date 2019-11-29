package cli

import (
	"testing"

	"github.com/elves/elvish/cli/el/codearea"
)

func TestCodeBufferUtils(t *testing.T) {
	app := NewApp(AppSpec{})
	buf := codearea.Buffer{Content: "test code", Dot: 2}
	SetCodeBuffer(app, buf)
	if gotBuf := CodeBuffer(app); gotBuf != buf {
		t.Errorf("Got buffer %v, want %v", gotBuf, buf)
	}
}
