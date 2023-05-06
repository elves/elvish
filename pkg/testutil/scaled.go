package testutil

import (
	"os"
	"strconv"
	"time"

	"src.elv.sh/pkg/env"
)

// Scaled returns d scaled by $E:ELVISH_TEST_TIME_SCALE. If the environment
// variable does not exist or contains an invalid value, the scale defaults to
// 1.
func Scaled(d time.Duration) time.Duration {
	return time.Duration(float64(d) * TestTimeScale())
}

// TestTimeScale parses $E:ELVISH_TEST_TIME_SCALE, defaulting to 1 it it's not
// set or can't be parsed as a float64.
func TestTimeScale() float64 {
	env := os.Getenv(env.ELVISH_TEST_TIME_SCALE)
	if env == "" {
		return 1
	}
	scale, err := strconv.ParseFloat(env, 64)
	if err != nil || scale <= 0 {
		return 1
	}
	return scale
}
