package main

import (
	"io"
	"os/exec"
)

type highlighter struct {
	Command *exec.Cmd
	Stdin   io.WriteCloser
	Writer  io.Writer
}

func NewHighlighter(w io.Writer) (*highlighter, error) {
	if !isTerminal() {
		return nil, nil
	}
	bat, err := exec.LookPath("bat")
	if err != nil {
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
