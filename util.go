package main

import (
	"os"

	"github.com/gechr/x/terminal"
)

// useColor reports whether output to stdout should be coloured: stdout must be
// a terminal, and NO_COLOR must be unset or empty (https://no-color.org/).
func useColor() bool {
	return os.Getenv("NO_COLOR") == "" && terminal.Is(os.Stdout)
}
