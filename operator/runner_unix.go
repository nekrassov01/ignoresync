//go:build !windows

package operator

import (
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

// Run executes the given command in the repository directory, forwarding
// signals to the child process group. The caller is responsible for pulling
// patterns and files beforehand and cleaning up afterwards.
func (o *Operator) Run(command string, environ []string) error {
	// Set up signal handling
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)

	// Resolve working directory
	dir := o.repo.path
	if o.workDir != "" && o.workDir != "." {
		dir = filepath.Join(o.repo.path, o.workDir)
	}

	// Resolve shell
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "sh"
	}
	cmd := exec.Command(sh, "-c", command) // #nosec G204,G702
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = environ
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return NewCommandError(err)
	}

	// Wait for the command to finish or a signal to be received
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case err := <-done:
		if err != nil {
			return NewCommandError(err)
		}
		return nil
	case sig := <-ch:
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal))
		}
		<-done
		return nil
	}
}
