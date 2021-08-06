package store_test

import (
	"testing"

	"src.elv.sh/pkg/store"
	"src.elv.sh/pkg/store/storetest"
)

func TestCmd(t *testing.T) {
	storetest.TestCmd(t, store.MustTempStore(t))
}
