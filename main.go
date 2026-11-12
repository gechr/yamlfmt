package main

import (
	"context"
	"errors"
	"fmt"
	"os"
)

func main() {
	err := run(parseFlags())
	if err != nil {
		if _, ok := errors.AsType[*signalError](err); !ok {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	os.Exit(exitCode(err))
}

func run(opts options) error {
	if opts.diff || opts.write {
		ctx, stop := notifyContext(context.Background())
		defer stop()
		return rewriteFiles(ctx, os.Stdout, opts)
	}
	var hl *highlighter
	if useColor() {
		hl = newHighlighter()
	}
	return printFiles(os.Stdout, os.Stdin, opts.paths, hl)
}

// exitCode returns the process exit status for the result of run.
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if sigErr, ok := errors.AsType[*signalError](err); ok {
		return sigErr.exitCode()
	}
	return 1
}
