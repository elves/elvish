// +build !elv_daemon_stub

package daemon

// No-op on Windows.
func setUmaskForDaemon() {}
