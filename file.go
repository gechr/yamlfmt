package main

import (
	"io"
	"os"
	"path/filepath"
)

const (
	fileSeparator = "---\n"
)

func filestream() *io.PipeReader {
	paths := os.Args[1:]
	if len(paths) == 0 || paths[0] == "-" {
		return nil
	}
	return streamfiles(paths)
}

/*
// Need to rethink this. Currently does not work because the `yaml.Decode`
// strips comments!
func header(w io.Writer, path, realpath string) {
	h := []byte("---\n# " + path)
    if path != realpath {
		h = append(h, " -> "+realpath...)
    }
    _, err := w.Write(h)
    errFatal(err)
}
*/

func streamfile(w io.Writer, path string) error {
	var err error
	_, err = io.WriteString(w, fileSeparator)
	if err != nil {
		return err
	}
	realpath := readlink(path)
	f, err := os.Open(realpath)
	if err != nil {
		errCh <- err
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}
	return nil
}

func streamfiles(paths []string) *io.PipeReader {
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		for _, path := range paths {
			errCh <- streamfile(w, path)
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
