package main

import (
	"context"
	"io"
	"os"
)

var errCh = make(chan error)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go run(cancel)
	exit(ctx)
}

func run(cancel context.CancelFunc) {
	f := NewFormatter()

	if flagWrite || flagDiff {
		errCh <- processfiles(f)
		cancel()
		return
	}

	if fs := filestream(); fs != nil {
		f.SetReader(fs)
	}

	h := NewHighlighter()

	// No highlighter, just stream to stdout and return.
	if h == nil {
		errCh <- f.Format()
		cancel()
		return
	}

	// Wire up the output of the formatter to the highlighter.
	reader, writer := io.Pipe()
	f.SetWriter(writer)
	h.SetReader(reader)

	go func() {
		defer writer.Close()
		errCh <- f.Format()
	}()

	errCh <- h.Highlight()
	cancel()
}

func exit(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			os.Exit(0)
		case err := <-errCh:
			if err != nil {
				fatal(err.Error())
			}
		}
	}
}
