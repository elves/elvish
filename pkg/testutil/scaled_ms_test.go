package testutil

import (
	"os"
	"testing"
	"time"
)

var scaledMsTests = []struct {
	name string
	env  string
	ms   int

	want time.Duration
}{
	{"default 10ms", "", 10, 10 * time.Millisecond},

	{"2x 10ms", "2", 10, 20 * time.Millisecond},
	{"2x 3000ms", "2", 3000, 6 * time.Second},
	{"0.5x 10ms", "0.5", 10, 5 * time.Millisecond},

	{"invalid treated as 1", "a", 10, 10 * time.Millisecond},
	{"0 treated as 1", "0", 10, 10 * time.Millisecond},
	{"negative treated as 1", "-1", 10, 10 * time.Millisecond},
}

func TestScaledMs(t *testing.T) {
	envSave := os.Getenv("ELVISH_TEST_TIME_SCALE")
	defer os.Setenv("ELVISH_TEST_TIME_SCALE", envSave)

	for _, test := range scaledMsTests {
		t.Run(test.name, func(t *testing.T) {
			os.Setenv("ELVISH_TEST_TIME_SCALE", test.env)
			got := ScaledMs(test.ms)
			if got != test.want {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}
