package rat

import (
	"io"
	"os"
	"os/exec"
	"syscall"
)

type ShellCommand interface {
	io.ReadCloser
}

type shellCommand struct {
	cmd *exec.Cmd
	io.Reader
}

func NewShellCommand(c string) (ShellCommand, error) {
	sc := &shellCommand{}

	sc.cmd = exec.Command(os.Getenv("SHELL"), "-c", c)
	sc.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var (
		stdout io.Reader
		stderr io.Reader
		err    error
	)

	if stdout, err = sc.cmd.StdoutPipe(); err != nil {
		return sc, err
	}

	if stderr, err = sc.cmd.StderrPipe(); err != nil {
		return sc, err
	}

	sc.Reader = io.MultiReader(stdout, stderr)

	err = sc.cmd.Start()

	return sc, err
}

func (sc *shellCommand) Close() error {
	sc.cmd.Wait()
	return syscall.Kill(-sc.cmd.Process.Pid, syscall.SIGTERM)
}
