package bundled

import "testing"

func TestGet(t *testing.T) {
	if Get()["epm"] != epmElv {
		t.Error(`Get()["epm"] != epmElv`)
	}
}
