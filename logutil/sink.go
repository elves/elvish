package logutil

import "log"

// Sink is a Logger that ignores all loggings.
var Sink = log.New(sinkWriter{}, "", 0)

type sinkWriter struct{}

func (sw sinkWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
