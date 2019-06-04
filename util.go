package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
)

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func isTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}
