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

func NewHighlighter(w io.Writer) *highlighter {
	if !isTerminal() {
		return nil
	}
	bat, err := exec.LookPath("bat")
	if err != nil {
		return nil
	}
	cmd := exec.Command(
		bat,
		"--color=always",
		"--language=yaml",
		"--paging=never",
		"--plain",
	)
	stdin, err := cmd.StdinPipe()
	errFatal(err)
	cmd.Stdout = w
	errFatal(cmd.Start())
	return &highlighter{
		Command: cmd,
		Stdin:   stdin,
		Writer:  w,
	}
}

func (h *highlighter) Highlight() error {
	return h.Command.Wait()
}
