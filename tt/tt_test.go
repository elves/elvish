package tt

import (
	"fmt"
	"testing"
)

// testT implements the T interface and is used to verify the Test function's
// interaction with T.
type testT []string

func (t *testT) Errorf(format string, args ...interface{}) {
	*t = append(*t, fmt.Sprintf(format, args...))
}

// fn is the function under test.
func fn(x int, y int) (int, int) {
	return x + y, x - y
}

func TestTTPass(t *testing.T) {
	var testT testT
	Test(&testT, Fn{"fn", "(%d, %d)", "(%d, %d)", fn}, Table{
		C(1, 10).Rets(11, -9),
	})
	if len(testT) > 0 {
		t.Errorf("Test errors when test should pass")
	}
}

func TestTTFail(t *testing.T) {
	var testT testT
	Test(&testT, Fn{"fn", "(%d, %d)", "(%d, %d)", fn}, Table{
		C(1, 10).Rets(11, -90),
	})
	switch len(testT) {
	case 0:
		t.Errorf("Test didn't error when it should")
	case 1:
		if testT[0] != "fn(1, 10) -> (11, -9), want (11, -90)" {
			t.Errorf("Test wrote message %q, not wanted", testT[0])
		}
	default:
		t.Errorf("Test wrote too many error messages")
	}
}
