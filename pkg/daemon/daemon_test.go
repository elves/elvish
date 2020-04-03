package daemon

import (
	"testing"
	"time"

	"github.com/elves/elvish/pkg/testutil"
	"github.com/elves/elvish/pkg/util"
)

func TestDaemon(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()

	serverDone := make(chan struct{})
	go func() {
		Serve("sock", "db")
		close(serverDone)
	}()

	client := NewClient("sock")
	for i := 0; i < 100; i++ {
		client.ResetConn()
		_, err := client.Version()
		if err == nil {
			break
		} else if i == 99 {
			t.Fatal("Failed to connect after 1s")
		}
		time.Sleep(testutil.ScaledMs(10))
	}

	_, err := client.AddCmd("test cmd")
	if err != nil {
		t.Errorf("client.AddCmd -> error %v", err)
	}
	client.Close()
	// Wait for server to quit before returning
	<-serverDone
}
