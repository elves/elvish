package store_test

import (
	"testing"

	"github.com/elves/elvish/pkg/store"
	"github.com/elves/elvish/pkg/store/storetest"
)

func TestCmd(t *testing.T) {
	tStore, cleanup := store.MustGetTempStore()
	defer cleanup()
	storetest.TestCmd(t, tStore)
}
