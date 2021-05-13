package testutil

import (
	"os"
	"strconv"
	"time"

	"src.elv.sh/pkg/env"
)

// ScaledMs returns ms milliseconds, scaled by the ELVISH_TEST_TIME_SCALE
// environment variable. If the variable does not exist, the scale defaults to
// 1.
func ScaledMs(ms int) time.Duration {
	return time.Duration(
		float64(ms) * float64(time.Millisecond) * getTestTimeScale())
}

func getTestTimeScale() float64 {
	env := os.Getenv(env.ElvishTestTimeScale)
	if env == "" {
		return 1
	}
	scale, err := strconv.ParseFloat(env, 64)
	if err != nil || scale <= 0 {
		return 1
	}
	return scale
}
