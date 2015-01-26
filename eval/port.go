package eval

import "os"

// A port conveys data stream. When f is not nil, it may convey fdStream. When
// ch is not nil, it may convey chanStream. When both are nil, it is always
// closed and may not convey any stream (unusedStream).
type port struct {
	f       *os.File
	ch      chan Value
	closeF  bool
	closeCh bool
}

// closePorts closes the suitable components of all ports in ev.ports that were
// marked marked for closing.
func (ev *Evaluator) closePorts() {
	for _, port := range ev.ports {
		if port.closeF {
			port.f.Close()
		}
		if port.closeCh {
			close(port.ch)
		}
	}
}
