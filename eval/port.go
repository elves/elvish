package eval

import "os"

// A port conveys data stream. When f is not nil, it may convey fdStream. When
// ch is not nil, it may convey chanStream. When both are nil, it is always
// closed and may not convey any stream.
type port struct {
	f       *os.File
	ch      chan Value
	closeF  bool
	closeCh bool
}

// close closes
func (p *port) close() {
	if p == nil {
		return
	}
	if p.closeF {
		p.f.Close()
	}
	if p.closeCh {
		close(p.ch)
	}
}

// closePorts closes a list of ports.
func closePorts(ports []*port) {
	for _, port := range ports {
		port.close()
	}
}

func (ec *evalCtx) closePorts() {
	closePorts(ec.ports)
}
