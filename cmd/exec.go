package cmd

import (
	"io"
	"os"
	"os/exec"
	"syscall"
)

type ReadKiller interface {
	io.Reader
	Kill() error
}

type readKiller struct {
	cmd *exec.Cmd
	rd  io.Reader
}

func Exec(command string) (ReadKiller, error) {
	var (
		r      *readKiller
		err    error
		stdout io.Reader
		stderr io.Reader
	)

	r = &readKiller{}

	r.cmd = exec.Command(os.Getenv("SHELL"), "-c", command)
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if stdout, err = r.cmd.StdoutPipe(); err != nil {
		return r, err
	}

	if stderr, err = r.cmd.StderrPipe(); err != nil {
		return r, err
	}

	r.rd = io.MultiReader(stdout, stderr)

	err = r.cmd.Start()


	return r, err
}

func (r *readKiller) Read(p []byte) (int, error) {
	return r.rd.Read(p)
}

func (r *readKiller) Kill() error {
	return syscall.Kill(-r.cmd.Process.Pid, syscall.SIGTERM)
}
