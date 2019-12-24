package cli_test

import (
	"testing"

	. "github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/el/codearea"
	"github.com/elves/elvish/pkg/cli/el/layout"
)

func TestCodeBufferUtils(t *testing.T) {
	app := NewApp(AppSpec{})
	buf := codearea.CodeBuffer{Content: "test code", Dot: 2}
	SetCodeBuffer(app, buf)
	if gotBuf := CodeBuffer(app); gotBuf != buf {
		t.Errorf("Got buffer %v, want %v", gotBuf, buf)
	}
}

func TestAddonUtils(t *testing.T) {
	app := NewApp(AppSpec{})
	addon := layout.Empty{}
	SetAddon(app, addon)
	if gotAddon := Addon(app); gotAddon != addon {
		t.Errorf("Got addon %v, want %v", gotAddon, addon)
	}
}
