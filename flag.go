package main

import (
	flag "github.com/spf13/pflag"
)

type options struct {
	diff  bool
	write bool
	paths []string
}

func parseFlags() options {
	var opts options
	flag.BoolVarP(&opts.diff, "diff", "d", false, "display diffs instead of rewriting files")
	flag.BoolVarP(
		&opts.write,
		"write",
		"w",
		false,
		"write result to (source) file instead of stdout",
	)
	flag.Parse()
	opts.paths = flag.Args()
	return opts
}
