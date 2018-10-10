package util

import (
	"reflect"
	"testing"
)

func TestFeed(t *testing.T) {
	var fed []interface{}

	Feed(func(x interface{}) bool {
		fed = append(fed, x)
		return x != 10
	}, 1, 2, 3, 10, 11, 12, 13)

	wantFed := []interface{}{1, 2, 3, 10}
	if !reflect.DeepEqual(fed, wantFed) {
		t.Errorf("Fed %v, want %v", fed, wantFed)
	}
}
