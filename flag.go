package main

import (
	flag "github.com/spf13/pflag"
)

var (
	// nolint:gochecknoglobals
	flagWrite bool
)

// nolint:gochecknoinits
func init() {
	flag.BoolVarP(&flagWrite, "write", "w", false, "write result to (source) file instead of stdout")
	flag.Parse()
}

func args() []string {
	return flag.Args()
}
