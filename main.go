package main

import (
	"context"
	"os"
)

var (
	errCh = make(chan error)
)

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

	stream := filestream()
	if stream != nil {
		f.SetReader(stream)
	}

	h, err := NewHighlighter(os.Stdout)
	if err != nil {
		errCh <- err
		cancel()
		return
	}

	// No highlighter, just stream to stdout and return
	if h == nil {
		errCh <- f.Format()
		cancel()
		return
	}

	// Pipe the output of the formatter to the highlighter
	f.SetWriter(h.Stdin)
	go func() {
		defer h.Stdin.Close()
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
