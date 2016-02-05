package logutil

import "log"

type sinkWriter struct{}

func (sw sinkWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

var Sink = log.New(sinkWriter{}, "", 0)
