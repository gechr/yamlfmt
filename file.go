package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	fileSeparator = "\n---\n"
)

func writefiles(f *formatter) error {
	paths := args()
	for _, path := range paths {
		// Input
		dst, err := os.Open(path)
		if err != nil {
			return err
		}
		f.SetReader(dst)
		defer dst.Close()
		// Output
		bak, err := ioutil.TempFile(filepath.Dir(path), filepath.Base(path)+".*")
		if err != nil {
			return err
		}
		f.SetWriter(bak)
		defer bak.Close()
		// Format
		if err := f.Format(); err != nil {
			return err
		}
		// Cleanup
		if err := os.Rename(bak.Name(), dst.Name()); err != nil {
			return err
		}
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
