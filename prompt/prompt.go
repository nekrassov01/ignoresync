package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/nekrassov01/ignoresync/color"
	"golang.org/x/term"
)

// Confirm prompts the user for confirmation with the given label.
// msg is expected to be a simple string such as "canceled" or
// "skipped" and will not be handled as an ignoresync error.
func Confirm(w io.Writer, label, msg string) (string, error) {
	_, _ = fmt.Fprintf(w, "%s %s ", label, color.Mute("[y/N]"))

	s, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", errors.New(msg)
	}

	s = strings.TrimSpace(s)
	if !strings.EqualFold(s, "y") {
		return s, errors.New(msg)
	}

	return s, nil
}

// Secret prompts the user to enter a secret value with the given label and validation function.
func Secret(w io.Writer, label string, validator func(string) error) (string, error) {
	r := bufio.NewReader(os.Stdin)
	for {
		_, _ = fmt.Fprintf(w, "%s ", label)

		secret, err := readSecret(os.Stdin, r)
		if err != nil {
			return "", NewPromptError(fmt.Errorf("failed to read string: %w", err))
		}

		if validator != nil {
			if err := validator(secret); err != nil {
				_, _ = fmt.Fprintln(w, color.Warn(err.Error()))
				continue
			}
		}
		_, _ = fmt.Fprintln(w)

		return secret, nil
	}
}

// readSecret reads a secret value from the given file.
func readSecret(f *os.File, r *bufio.Reader) (string, error) {
	if isTerminal(f) {
		return readTerminalSecret(f)
	}
	if r == nil {
		r = bufio.NewReader(f)
	}
	s, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(s, "\r\n"), nil
}

// readTerminalSecret reads a secret value from the terminal without echo.
func readTerminalSecret(f *os.File) (string, error) {
	p := f.Fd()
	if p > uintptr(math.MaxInt) {
		return "", fmt.Errorf("file descriptor out of range: %d", p)
	}
	fd := int(p)

	state, err := term.GetState(fd)
	if err != nil {
		return "", fmt.Errorf("failed to get terminal state: %w", err)
	}

	// Setup signal handler to restore terminal state on interrupt
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	defer func() {
		signal.Stop(ch)
		if state != nil {
			_ = term.Restore(fd, state)
		}
	}()

	// If an interrupt is received,
	// restore terminal and exit to avoid leaving echo off
	go func() {
		<-ch
		if state != nil {
			_ = term.Restore(fd, state)
		}
		os.Exit(1)
	}()

	// Read secret without echo
	secret, err := term.ReadPassword(fd)
	if err != nil {
		if state != nil {
			_ = term.Restore(fd, state)
		}
		return "", err
	}

	// Ensure terminal restored after successful read
	if state != nil {
		_ = term.Restore(fd, state)
	}

	return string(secret), nil
}

// isTerminal checks if the given file is a terminal.
func isTerminal(f *os.File) bool {
	fd := f.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
