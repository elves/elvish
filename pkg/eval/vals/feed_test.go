package vals

import (
	"reflect"
	"testing"
)

func TestFeed(t *testing.T) {
	var fed []any

	Feed(func(x any) bool {
		fed = append(fed, x)
		return x != 10
	}, 1, 2, 3, 10, 11, 12, 13)

	wantFed := []any{1, 2, 3, 10}
	if !reflect.DeepEqual(fed, wantFed) {
		t.Errorf("Fed %v, want %v", fed, wantFed)
	}
}
