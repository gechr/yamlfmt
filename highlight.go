package main

import (
	"io"
	"os"
	"os/exec"
)

type highlighter struct {
	Command *exec.Cmd
	Stdin   io.WriteCloser
	Writer  io.Writer
}

func shouldHighlight() bool {
	// If NO_COLOR is set, do not highlight (https://no-color.org/)
	if _, nc := os.LookupEnv("NO_COLOR"); nc {
		return false
	}
	// If not attached to a terminal, do not highlight
	if !isTerminal() {
		return false
	}
	return true
}

func findBat() string {
	// If `bat` is not on the user's path, do not highlight
	bat, err := exec.LookPath("bat")
	if err != nil {
		return ""
	}
	return bat
}

func NewHighlighter(w io.Writer) (*highlighter, error) {
	if !shouldHighlight() {
		return nil, nil
	}
	bat := findBat()
	if bat == "" {
		return nil, nil
	}
	cmd := exec.Command(
		bat,
		"--color=always",
		"--language=yaml",
		"--paging=never",
		"--plain",
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stdout = w
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &highlighter{
		Command: cmd,
		Stdin:   stdin,
		Writer:  w,
	}, nil
}

func (h *highlighter) Highlight() error {
	return h.Command.Wait()
}
