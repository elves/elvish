package eval

import "os"

// Port conveys data stream. It always consists of a byte band and a channel band.
type Port struct {
	File      *os.File
	Chan      chan Value
	CloseFile bool
	CloseChan bool
}

// close closes a Port.
func (p *Port) close() {
	if p == nil {
		return
	}
	if p.CloseFile {
		p.File.Close()
	}
	if p.CloseChan {
		close(p.Chan)
	}
}

// closePorts closes a list of Ports.
func closePorts(ports []*Port) {
	for _, port := range ports {
		port.close()
	}
}
