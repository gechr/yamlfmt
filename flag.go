package main

import (
	flag "github.com/spf13/pflag"
)

var (
	flagDiff  bool
	flagWrite bool
)

func init() {
	flag.BoolVarP(&flagDiff, "diff", "d", false, "display diffs instead of rewriting files")
	flag.BoolVarP(&flagWrite, "write", "w", false, "write result to (source) file instead of stdout")
	flag.Parse()
}

func args() []string {
	return flag.Args()
}
