package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

var (
	bakName string

	sigs = make(chan os.Signal, 1)
)

const (
	fileSeparator = "\n---\n"
)

func diff(before, after string) error {
	data, err := exec.Command(
		"git",
		"diff",
		"--no-index",
		"--color=always",
		"--",
		before,
		after,
	).CombinedOutput()
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		fmt.Print(string(data))
		return nil
	}
	return err
}

func backupCleaner() {
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		if bakName != "" {
			_ = os.Remove(bakName)
		}
		os.Exit(0)
	}()
}

func processfiles(f *formatter) error {
	backupCleaner()
	paths := args()
	if len(paths) == 0 || paths[0] == "-" {
		return errors.New(`cannot use "--diff" or "--write" when reading from stdin`)
	}
	for _, path := range paths {
		// Input
		dst, err := os.Open(path)
		if err != nil {
			return errors.Wrapf(err, "failed to read file [%s]", path)
		}
		f.SetReader(dst)
		defer dst.Close()
		// Output
		bak, err := ioutil.TempFile(filepath.Dir(path), filepath.Base(path)+".*")
		bakName = bak.Name()
		if err != nil {
			return errors.Wrap(err, "failed to create temp file")
		}
		f.SetWriter(bak)
		defer bak.Close()
		// Format
		if err := f.Format(); err != nil {
			return errors.Wrapf(err, "failed to format file [%s]", path)
		}
		// Diff
		if flagDiff {
			if err := diff(dst.Name(), bakName); err != nil {
				return errors.Wrapf(err, "failed to diff file [%s] with [%s]", dst.Name(), bakName)
			}
			if !flagWrite {
				if err := os.Remove(bak.Name()); err != nil {
					return errors.Wrapf(err, "failed to delete file [%s]", bak.Name())
				}
			}
		}
		// Write
		if flagWrite {
			if err := os.Rename(bakName, dst.Name()); err != nil {
				return errors.Wrapf(err, "failed to rename file [%s] to [%s]", bakName, dst.Name())
			}
		}
		bakName = ""
	}
	return nil
}

func filestream() *io.PipeReader {
	paths := args()
	if len(paths) == 0 || paths[0] == "-" {
		return nil
	}
	return streamfiles(paths)
}

func writeHeader(w io.Writer, path, realpath string) error {
	header := []byte("# " + path)
	if path != realpath {
		symlink := " -> " + realpath
		header = append(header, symlink...)
	}
	header = append(header, fileSeparator...)
	_, err := w.Write(header)
	return err
}

func streamfile(w io.Writer, path string, includeHeader bool) error {
	realpath := readlink(path)
	if includeHeader {
		if err := writeHeader(w, path, realpath); err != nil {
			return err
		}
	}
	f, err := os.Open(realpath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(w, f); err != nil {
		return err
	}
	return nil
}

func streamfiles(paths []string) *io.PipeReader {
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		includeHeaders := len(paths) > 1
		for _, path := range paths {
			errCh <- streamfile(w, path, includeHeaders)
		}
	}()
	return r
}

func readlink(path string) string {
	p, _ := filepath.EvalSymlinks(path)
	if p == "" {
		return path
	}
	return p
}
