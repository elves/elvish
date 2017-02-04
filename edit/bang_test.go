package edit

import (
	"reflect"
	"testing"
)

var (
	theLine = "qw search 'foo bar ~y'"
	theBang = newBang(theLine)

	wantLen   = 4
	wantShown = []styled{
		unstyled("M-, " + theLine),
		unstyled("  0 qw"),
		unstyled("  1 search"),
		unstyled("  2 'foo bar ~y'"),
	}
	wantShownMinus = []styled{
		unstyled("M-, " + theLine),
		unstyled(" -3 qw"),
		unstyled(" -2 search"),
		unstyled(" -1 'foo bar ~y'"),
	}
)

func TestBang(t *testing.T) {
	if len := theBang.Len(); len != wantLen {
		t.Errorf("theBang.Len() -> %d, want 3", len)
	}

	// Test the result of Show for all indices.
	testShow := func(wants []styled) {
		for i, want := range wants {
			if shown := theBang.Show(i); !reflect.DeepEqual(shown, want) {
				t.Errorf("theBang.Show(%d) => %v, want %v", i, shown, want)
			}
		}
	}

	testShow(wantShown)

	theBang.Filter("-")
	testShow(wantShownMinus)

	// Test that there is only one item, and Show(0) matches want.
	testOne := func(want styled) {
		if len := theBang.Len(); len != 1 {
			t.Errorf("theBang.Len() -> %d, want 1", len)
		}
		if shown := theBang.Show(0); !reflect.DeepEqual(shown, want) {
			t.Errorf("theBang.Show(0) => %v, want %v", shown, want)
		}
	}

	theBang.Filter("1")
	testOne(wantShown[2])
	theBang.Filter("-1")
	testOne(wantShownMinus[len(wantShownMinus)-1])
}
