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
	go run(ctx, cancel)
	exit(ctx)
}

func run(ctx context.Context, cancel context.CancelFunc) {
	f := NewFormatter()
	stream := filestream()
	if stream != nil {
		f.SetReader(stream)
	}

	h, err := NewHighlighter(os.Stdout)
	if err != nil {
		errCh <- err
		return
	}
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
