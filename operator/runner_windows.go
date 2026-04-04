//go:build windows

package operator

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"
)

// Run executes the given command in the repository directory. On Windows,
// signals are handled by the OS and forwarded to the child process
// automatically. The caller is responsible for pulling patterns and files
// beforehand and cleaning up afterwards.
func (o *Operator) Run(ctx context.Context, command string, environ []string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	dir := o.repo.path
	if o.workDir != "" && o.workDir != "." {
		dir = filepath.Join(o.repo.path, o.workDir)
	}

	cmd := exec.CommandContext(ctx, "cmd", "/c", command) // #nosec G204,G702
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = environ

	cmd.Cancel = func() error {
		if cmd.Process != nil {
			return cmd.Process.Signal(os.Interrupt)
		}
		return nil
	}
	cmd.WaitDelay = 5 * time.Second

	if err := cmd.Start(); err != nil {
		return NewCommandError(err)
	}

	if err := cmd.Wait(); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil
		}
		return NewCommandError(err)
	}

	return nil
}
