package bc

import (
	"io"
	"log"
	"os"
	"os/exec"
)

type Bc interface {
	Exec(string) error
	Quit()
}

type bc struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func Start() Bc {
	cmd := exec.Command("bc")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	cmd.Start()
	return &bc{cmd, stdin, stdout}
}

// TODO: Use a more robust marker
var inputSuffix = []byte("\n\"\x04\"\n")

func (bc *bc) Exec(code string) error {
	bc.stdin.Write([]byte(code))
	bc.stdin.Write(inputSuffix)
	for {
		b, err := readByte(bc.stdout)
		if err != nil {
			return err
		}
		if b == 0x04 {
			break
		}
		os.Stdout.Write([]byte{b})
	}
	return nil
}

func readByte(r io.Reader) (byte, error) {
	var buf [1]byte
	_, err := r.Read(buf[:])
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (bc *bc) Quit() {
	bc.stdin.Close()
	bc.cmd.Wait()
	bc.stdout.Close()
}
