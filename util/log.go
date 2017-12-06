package util

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	out     io.Writer = ioutil.Discard
	logFile *os.File
	loggers []*log.Logger
)

// GetLogger gets a logger with a prefix.
func GetLogger(prefix string) *log.Logger {
	logger := log.New(out, prefix, log.LstdFlags)
	loggers = append(loggers, logger)
	return logger
}

func setOutput(newout io.Writer) {
	out = newout
	for _, logger := range loggers {
		logger.SetOutput(out)
	}
}

func SetOutputFile(fname string) error {
	if logFile != nil {
		logFile.Close()
	}
	if fname == "" {
		logFile = nil
		setOutput(ioutil.Discard)
		return nil
	}
	var err error
	logFile, err = os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	setOutput(logFile)
	return nil
}
