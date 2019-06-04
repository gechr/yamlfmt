package main

import (
	"os"
)

func main() {
	f := NewFormatter()
	stream := filestream()
	if stream != nil {
		f.SetReader(stream)
	}

	h := NewHighlighter(os.Stdout)
	if h == nil {
		errFatal(f.Format())
		return
	}

	f.SetWriter(h.Stdin)
	go func() {
		defer h.Stdin.Close()
		errFatal(f.Format())
	}()

	errFatal(h.Highlight())
}
