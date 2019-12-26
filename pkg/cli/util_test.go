package cli_test

import (
	"testing"

	. "github.com/elves/elvish/pkg/cli"
)

func TestCodeBufferUtils(t *testing.T) {
	app := NewApp(AppSpec{})
	buf := CodeBuffer{Content: "test code", Dot: 2}
	SetCodeBuffer(app, buf)
	if gotBuf := GetCodeBuffer(app); gotBuf != buf {
		t.Errorf("Got buffer %v, want %v", gotBuf, buf)
	}
}

func TestAddonUtils(t *testing.T) {
	app := NewApp(AppSpec{})
	addon := Empty{}
	SetAddon(app, addon)
	if gotAddon := Addon(app); gotAddon != addon {
		t.Errorf("Got addon %v, want %v", gotAddon, addon)
	}
}
