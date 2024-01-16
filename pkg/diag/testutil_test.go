package diag

import (
	"testing"

	"src.elv.sh/pkg/testutil"
)

var dedent = testutil.Dedent

func setContextBodyMarkers(t *testing.T, start, end string) {
	testutil.Set(t, &ContextBodyStartMarker, start)
	testutil.Set(t, &ContextBodyEndMarker, end)
}

func setMessageMarkers(t *testing.T, start, end string) {
	testutil.Set(t, &messageStart, start)
	testutil.Set(t, &messageEnd, end)
}
