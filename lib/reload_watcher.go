package rat

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type ReloadWatcher interface {
	Stop() error
}

type reloadWatcher struct {
	cmd *exec.Cmd
}

func NewReloadWatcher(p Pager, cmd string, ctx Context) (ReloadWatcher, error) {
	rw := &reloadWatcher{}

	rw.cmd = exec.Command(os.Getenv("SHELL"), "-c", cmd)
	rw.cmd.Env = ContextEnvironment(ctx)
	rw.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var (
		stdout io.Reader
		err    error
	)

	if stdout, err = rw.cmd.StdoutPipe(); err != nil {
		return rw, err
	}

	err = rw.cmd.Start()
	go func() { rw.cmd.Wait() }()

	go rw.scanLines(stdout, p)

	return rw, err
}

func (rw *reloadWatcher) scanLines(rd io.Reader, p Pager) {
	scanner := bufio.NewScanner(rd)

	var timer *time.Timer

	for scanner.Scan() {
		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(500*time.Millisecond, p.Reload)
	}
}

func (rw *reloadWatcher) Stop() error {
	err := syscall.Kill(-rw.cmd.Process.Pid, syscall.SIGTERM)
	rw.cmd.Wait()
	return err
}
