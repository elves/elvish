package eval

import (
	"testing"

	"github.com/elves/elvish/pkg/parse"
)

func BenchmarkOutputCaptureOverhead(b *testing.B) {
	benchmarkOutputCapture(b.N, func(fm *Frame) {})
}

func BenchmarkOutputCaptureValues(b *testing.B) {
	benchmarkOutputCapture(b.N, func(fm *Frame) {
		fm.ports[1].Chan <- "test"
	})
}

func BenchmarkOutputCaptureBytes(b *testing.B) {
	bytesToWrite := []byte("test")
	benchmarkOutputCapture(b.N, func(fm *Frame) {
		fm.ports[1].File.Write(bytesToWrite)
	})
}

func BenchmarkOutputCaptureMixed(b *testing.B) {
	bytesToWrite := []byte("test")
	benchmarkOutputCapture(b.N, func(fm *Frame) {
		fm.ports[1].Chan <- false
		fm.ports[1].File.Write(bytesToWrite)
	})
}

func benchmarkOutputCapture(n int, f func(*Frame)) {
	ev := NewEvaler()
	defer ev.Close()
	fm := NewTopFrame(ev, parse.Source{Name: "[benchmark]"}, []*Port{{}, {}, {}})
	for i := 0; i < n; i++ {
		captureOutput(fm, func(fm *Frame) error {
			f(fm)
			return nil
		})
	}
}
