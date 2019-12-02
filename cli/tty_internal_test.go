package cli

import (
	"os"
	"testing"
	"time"
)

func TestGetUITestTimeout(t *testing.T) {
	original := os.Getenv(uiTimeoutEnvName)
	defer os.Setenv(uiTimeoutEnvName, original)

	os.Unsetenv(uiTimeoutEnvName)
	timeout := getUITestTimeout()
	if timeout != uiTimeoutDefault {
		t.Errorf("Not default when env not set")
	}

	os.Setenv(uiTimeoutEnvName, "10s")
	timeout = getUITestTimeout()
	if timeout != 10*time.Second {
		t.Errorf("Not set from environment variable")
	}
}
