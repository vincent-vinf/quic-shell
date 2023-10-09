package shell

import (
	"io"
	"log"
	"os/exec"

	"github.com/creack/pty"
)

func New(in io.Reader, out io.Writer) error {
	c := exec.Command("bash")
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	defer func() { _ = ptmx.Close() }()

	go func() {
		_, _ = io.Copy(ptmx, in)
		log.Println("shell ptmx out end")
	}()
	_, _ = io.Copy(out, ptmx)
	log.Println("shell ptmx out end")
	return nil
}
