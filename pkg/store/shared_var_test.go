package store_test

import (
	"testing"

	"src.elv.sh/pkg/store"
	"src.elv.sh/pkg/store/storetest"
)

func TestSharedVar(t *testing.T) {
	storetest.TestSharedVar(t, store.MustTempStore(t))
}
