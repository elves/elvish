package store_test

import (
	"testing"

	"src.elv.sh/pkg/store"
	"src.elv.sh/pkg/store/storetest"
)

func TestDir(t *testing.T) {
	storetest.TestDir(t, store.MustTempStore(t))
}
