package main

import (
	"io"
	"os"
	"path/filepath"
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

func streamfiles(paths []string) *io.PipeReader {
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		var err error
		for _, path := range paths {
			realpath := readlink(path)
			_, err = io.WriteString(w, "---\n")
			errFatal(err)
			f, err := os.Open(realpath)
			errFatal(err)
			_, err = io.Copy(w, f)
			errFatal(err)
			f.Close()
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
