package cli_test

import (
	"testing"

	. "src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/tk"
)

func TestCodeBufferUtils(t *testing.T) {
	app := NewApp(AppSpec{})
	buf := tk.CodeBuffer{Content: "test code", Dot: 2}
	SetCodeBuffer(app, buf)
	if gotBuf := CodeBuffer(app); gotBuf != buf {
		t.Errorf("Got buffer %v, want %v", gotBuf, buf)
	}
}

func TestAddonUtils(t *testing.T) {
	app := NewApp(AppSpec{})
	addon := tk.Empty{}
	SetAddon(app, addon)
	if gotAddon := Addon(app); gotAddon != addon {
		t.Errorf("Got addon %v, want %v", gotAddon, addon)
	}
}
