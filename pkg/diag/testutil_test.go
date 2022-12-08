package diag

import (
	"testing"

	"src.elv.sh/pkg/testutil"
)

var dedent = testutil.Dedent

func setCulpritMarkers(t *testing.T, start, end string) {
	testutil.Set(t, &culpritStart, start)
	testutil.Set(t, &culpritEnd, end)
}

func setMessageMarkers(t *testing.T, start, end string) {
	testutil.Set(t, &messageStart, start)
	testutil.Set(t, &messageEnd, end)
}
