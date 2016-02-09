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
func (p *Port) Close() {
	if p == nil {
		return
	}
	if p.CloseFile {
		p.File.Close()
	}
	if p.CloseChan {
		// Logger.Printf("closing channel %v", p.Chan)
		close(p.Chan)
	}
}

// closePorts closes a list of Ports.
func ClosePorts(ports []*Port) {
	for _, port := range ports {
		// Logger.Printf("closing port %d", i)
		port.Close()
	}
}
