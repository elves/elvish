package store_test

import (
	"testing"

	"src.elv.sh/pkg/store"
	"src.elv.sh/pkg/store/storetest"
)

func TestSharedVar(t *testing.T) {
	tStore, cleanup := store.MustGetTempStore()
	defer cleanup()
	storetest.TestSharedVar(t, tStore)
}
