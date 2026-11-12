package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	xslices "github.com/gechr/x/slices"
)

const stdinPath = "-"

var errStdinTwice = errors.New("standard input can only be read once")

// printFiles formats each input to w, reading stdin for "-" or when there are
// no paths. A single input streams to w one document at a time. Several inputs
// are formatted concurrently and written in order as one YAML stream, each
// introduced by a comment naming it. printFiles keeps going after an error and
// returns all of them.
func printFiles(w io.Writer, stdin io.Reader, paths []string, hl *highlighter) error {
	if len(paths) == 0 {
		paths = []string{stdinPath}
	}
	if len(paths) == 1 {
		if err := printFile(stdin, paths[0], false, emitTo(w, hl)); err != nil {
			return fmt.Errorf("%s: %w", displayName(paths[0]), err)
		}
		return nil
	}
	if xslices.Count(paths, stdinPath) > 1 {
		return errStdinTwice
	}
	type result struct {
		out []byte
		err error
	}
	var (
		errs     []error
		writeErr error
	)
	forEachOrdered(paths, workers(), func(path string) result {
		var buf bytes.Buffer
		err := printFile(stdin, path, true, emitTo(&buf, hl))
		// Keep the documents formatted before any error, as a single input
		// would have streamed them.
		return result{out: buf.Bytes(), err: err}
	}, func(path string, r result) {
		if writeErr == nil {
			_, writeErr = w.Write(r.out)
		}
		if r.err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", displayName(path), r.err))
		}
	})
	if writeErr != nil {
		errs = append(errs, writeErr)
	}
	return errors.Join(errs...)
}

func printFile(stdin io.Reader, path string, withHeader bool, emit func([]byte) error) error {
	if path == stdinPath {
		if withHeader {
			if err := emit(header(displayName(path))); err != nil {
				return err
			}
		}
		return format(stdin, formatOptions{continued: withHeader}, emit)
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if withHeader {
		name := path
		if target, err := filepath.EvalSymlinks(path); err == nil && target != path {
			name += " -> " + target
		}
		if err := emit(header(name)); err != nil {
			return err
		}
	}
	return format(f, formatOptions{continued: withHeader}, emit)
}

func emitTo(w io.Writer, hl *highlighter) func([]byte) error {
	if hl != nil {
		return func(b []byte) error {
			return hl.write(w, b)
		}
	}
	return func(b []byte) error {
		_, err := w.Write(b)
		return err
	}
}

// header returns the comment that introduces an input. A name with
// characters that could end the comment line is quoted.
func header(name string) []byte {
	if quoted := strconv.Quote(name); quoted[1:len(quoted)-1] != name {
		name = quoted
	}
	return []byte("# " + name + "\n")
}

func displayName(path string) string {
	if path == stdinPath {
		return "<stdin>"
	}
	return path
}
