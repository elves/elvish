package testutil

import (
	"os"
	"testing"
	"time"

	"src.elv.sh/pkg/env"
)

var scaledMsTests = []struct {
	name string
	env  string
	d    time.Duration

	want time.Duration
}{
	{"default 10ms", "", 10 * time.Millisecond, 10 * time.Millisecond},

	{"2x 10ms", "2", 10 * time.Millisecond, 20 * time.Millisecond},
	{"2x 3s", "2", 3 * time.Second, 6 * time.Second},
	{"0.5x 10ms", "0.5", 10 * time.Millisecond, 5 * time.Millisecond},

	{"invalid treated as 1", "a", 10 * time.Millisecond, 10 * time.Millisecond},
	{"0 treated as 1", "0", 10 * time.Millisecond, 10 * time.Millisecond},
	{"negative treated as 1", "-1", 10 * time.Millisecond, 10 * time.Millisecond},
}

func TestScaled(t *testing.T) {
	envSave := os.Getenv(env.ELVISH_TEST_TIME_SCALE)
	defer os.Setenv(env.ELVISH_TEST_TIME_SCALE, envSave)

	for _, test := range scaledMsTests {
		t.Run(test.name, func(t *testing.T) {
			os.Setenv(env.ELVISH_TEST_TIME_SCALE, test.env)
			got := Scaled(test.d)
			if got != test.want {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}
