package logutil

import (
	"io/ioutil"
	"log"
)

// Discard is a Logger that ignores all loggings.
var Discard = log.New(ioutil.Discard, "", 0)
