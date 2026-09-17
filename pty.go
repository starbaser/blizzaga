package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/caarlos0/go-shellwords"
	"github.com/charmbracelet/x/term"
	"github.com/charmbracelet/x/xpty"
)

func executeCommand(config Config) (string, error) {
	args, err := shellwords.Parse(config.Execute)
	if err != nil {
		return "", fmt.Errorf("could not execute: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.ExecuteTimeout)
	defer cancel()

	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		width = 80
		height = 24
	}

	pty, err := xpty.NewPty(width, height)
	if err != nil {
		return "", fmt.Errorf("could not execute: %w", err)
	}
	defer func() { _ = pty.Close() }()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...) //nolint: gosec
	if err := pty.Start(cmd); err != nil {
		return "", fmt.Errorf("could not execute: %w", err)
	}

	var out bytes.Buffer
	copied := make(chan struct{})
	go func() {
		defer close(copied)
		_, _ = io.Copy(&out, pty)
	}()

	waitErr := xpty.WaitProcess(ctx, cmd)
	// The copy ends when the child's side of the terminal closes. Reading the
	// buffer before then races the copy, so wait for it; if a grandchild still
	// holds the terminal past the deadline, close our side to end the copy.
	select {
	case <-copied:
	case <-ctx.Done():
		_ = pty.Close()
		<-copied
	}
	if waitErr != nil {
		return out.String(), fmt.Errorf("could not execute: %w", waitErr)
	}
	return out.String(), nil
}
