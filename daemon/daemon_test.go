package daemon

import (
	"testing"
	"time"

	"github.com/elves/elvish/util"
)

func TestDaemon(t *testing.T) {
	util.InTempDir(func(string) {
		serverDone := make(chan struct{})
		go func() {
			Serve("sock", "db")
			close(serverDone)
		}()

		client := NewClient("sock")
		for i := 0; i < 10; i++ {
			client.ResetConn()
			_, err := client.Version()
			if err == nil {
				break
			} else if i == 9 {
				t.Fatal("Failed to connect after 100ms")
			}
			time.Sleep(10 * time.Millisecond)
		}
		_, err := client.AddCmd("test cmd")
		if err != nil {
			t.Errorf("client.AddCmd -> error %v", err)
		}
		client.Close()
		// Wait for server to quit before returning
		<-serverDone
	})
}
