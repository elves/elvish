package eval

import (
	"fmt"
	"reflect"
)

// Common testing utilities.

// compareValues compares two slices, using equals for each element.
func compareSlice(wantValues, gotValues []interface{}) error {
	if len(wantValues) != len(gotValues) {
		return fmt.Errorf("want %d values, got %d",
			len(wantValues), len(gotValues))
	}
	for i, want := range wantValues {
		if !equals(want, gotValues[i]) {
			return fmt.Errorf("want [%d] = %s, got %s", i, want, gotValues[i])
		}
	}
	return nil
}

// equals compares two values. It uses Eq if want is a Value instance, or
// reflect.DeepEqual otherwise.
func equals(a, b interface{}) bool {
	if aValue, ok := a.(Value); ok {
		return aValue.Equal(b)
	}
	return reflect.DeepEqual(a, b)
}
