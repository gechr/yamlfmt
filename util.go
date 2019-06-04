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

// nolint:deadcode,unused
func fatalf(format string, a ...interface{}) {
	fatal(fmt.Sprintf(format, a...))
	os.Exit(1)
}

func errFatal(err error) {
	if err != nil {
		fatal(err.Error())
	}
}

func isTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}