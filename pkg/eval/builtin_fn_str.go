package eval

import (
	"fmt"
	"math"
	"math/big"
	"strconv"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/wcwidth"
)

// String operations.

// TODO(xiaq): Document -override-wcswidth.

func init() {
	addBuiltinFns(map[string]any{
		"<s":  chainStringComparer(func(a, b string) bool { return a < b }),
		"<=s": chainStringComparer(func(a, b string) bool { return a <= b }),
		"==s": chainStringComparer(func(a, b string) bool { return a == b }),
		">s":  chainStringComparer(func(a, b string) bool { return a > b }),
		">=s": chainStringComparer(func(a, b string) bool { return a >= b }),
		"!=s": func(a, b string) bool { return a != b },

		"to-string": toString,

		"base": base,

		"wcswidth":          wcwidth.Of,
		"-override-wcwidth": wcwidth.Override,
	})
}

func chainStringComparer(p func(a, b string) bool) func(...string) bool {
	return func(s ...string) bool {
		for i := 0; i < len(s)-1; i++ {
			if !p(s[i], s[i+1]) {
				return false
			}
		}
		return true
	}
}

func toString(fm *Frame, args ...any) error {
	out := fm.ValueOutput()
	for _, a := range args {
		err := out.Put(vals.ToString(a))
		if err != nil {
			return err
		}
	}
	return nil
}

func base(fm *Frame, b int, nums ...vals.Num) error {
	if b < 2 || b > 36 {
		return errs.OutOfRange{What: "base",
			ValidLow: "2", ValidHigh: "36", Actual: strconv.Itoa(b)}
	}
	// Don't output anything yet in case some arguments are invalid.
	results := make([]string, len(nums))
	for i, num := range nums {
		switch num := num.(type) {
		case int:
			results[i] = strconv.FormatInt(int64(num), b)
		case *big.Int:
			results[i] = num.Text(b)
		case float64:
			if i64 := int64(num); float64(i64) == num {
				results[i] = strconv.FormatInt(i64, b)
			} else if num == math.Trunc(num) {
				var z big.Int
				z.SetString(fmt.Sprintf("%.0f", num), 10)
				results[i] = z.Text(b)
			} else {
				return errs.BadValue{What: "number",
					Valid: "integer", Actual: vals.ReprPlain(num)}
			}
		default:
			return errs.BadValue{What: "number",
				Valid: "integer", Actual: vals.ReprPlain(num)}
		}
	}

	out := fm.ValueOutput()
	for _, s := range results {
		err := out.Put(s)
		if err != nil {
			return err
		}
	}
	return nil
}
