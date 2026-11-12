package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	xfilepath "github.com/gechr/x/filepath"
	xos "github.com/gechr/x/os"
	xslices "github.com/gechr/x/slices"
)

// signalExitBase is the shell convention for the exit status of a process
// stopped by a signal: 128 plus the signal number.
const signalExitBase = 128

var (
	errNotRegular = errors.New("not a regular file")
	errStdin      = errors.New("--diff and --write need file arguments, not standard input")
)

// signalError records which signal interrupted a run.
type signalError struct {
	sig syscall.Signal
}

func (e *signalError) Error() string {
	return e.sig.String() + " signal received"
}

func (e *signalError) exitCode() int {
	return signalExitBase + int(e.sig)
}

// rewriteJob is one file to rewrite: path as given, and the absolute path of
// the file it resolves to through any symlinks.
type rewriteJob struct {
	path   string
	target string
	err    error
}

// rewriteFiles formats files in place (--write) or shows how they would
// change (--diff), several at a time. Diffs and errors are written in the
// order the files were given. Cancelling ctx stops new files from starting,
// and never leaves a file half written.
func rewriteFiles(ctx context.Context, stdout io.Writer, opts options) error {
	if len(opts.paths) == 0 || xslices.ContainsAny(opts.paths, stdinPath) {
		return errStdin
	}
	type result struct {
		diff []byte
		err  error
	}
	var (
		errs     []error
		writeErr error
	)
	forEachOrdered(resolveJobs(opts.paths), workers(), func(job rewriteJob) result {
		if job.err != nil {
			return result{err: job.err}
		}
		if ctx.Err() != nil {
			return result{err: context.Cause(ctx)}
		}
		var diff bytes.Buffer
		err := rewriteFile(ctx, job, opts, &diff)
		return result{diff: diff.Bytes(), err: err}
	}, func(job rewriteJob, r result) {
		if writeErr == nil {
			_, writeErr = stdout.Write(r.diff)
		}
		if r.err != nil && ctx.Err() == nil {
			errs = append(errs, fmt.Errorf("%s: %w", job.path, r.err))
		}
	})
	if ctx.Err() != nil {
		return context.Cause(ctx)
	}
	if writeErr != nil {
		errs = append(errs, writeErr)
	}
	return errors.Join(errs...)
}

// resolveJobs resolves each path through symlinks, and drops any path that
// leads to a file already listed, so that no two workers replace one file.
func resolveJobs(paths []string) []rewriteJob {
	jobs := make([]rewriteJob, 0, len(paths))
	for _, path := range paths {
		target, err := xfilepath.Resolve(path)
		jobs = append(jobs, rewriteJob{path: path, target: target, err: err})
	}
	return xslices.UniqueFunc(jobs, func(job rewriteJob) string {
		if job.err != nil {
			return "\x00" + job.path
		}
		return job.target
	})
}

func rewriteFile(ctx context.Context, job rewriteJob, opts options, diffOut io.Writer) error {
	info, err := os.Stat(job.target)
	if err != nil {
		return err
	}
	// Reading a FIFO or device would block or never end, and renaming over
	// one would replace it.
	if !info.Mode().IsRegular() {
		return errNotRegular
	}
	src, err := os.ReadFile(job.target)
	if err != nil {
		return err
	}
	formatted, err := formatBytes(src)
	if err != nil {
		return err
	}
	if bytes.Equal(src, formatted) {
		return nil
	}
	if opts.diff {
		if err := showDiff(ctx, diffOut, diffLabel(job), formatted); err != nil {
			return err
		}
	}
	if !opts.write {
		return nil
	}
	if ctx.Err() != nil {
		return context.Cause(ctx)
	}
	// Formatted output can be produced again, so skip the fsyncs, which cost
	// milliseconds per file on macOS.
	return xos.AtomicWrite(job.target, formatted, info.Mode().Perm(), xos.WithNoSync())
}

// diffLabel returns the path git should show for job: the path as given,
// unless it is a symlink, which git would compare as a link.
func diffLabel(job rewriteJob) string {
	if link, err := xos.IsSymlink(job.path); err == nil && !link {
		return job.path
	}
	return job.target
}

// showDiff runs git to write the difference between the file at path and
// formatted to w.
func showDiff(ctx context.Context, w io.Writer, path string, formatted []byte) error {
	tmp, err := os.CreateTemp("", "yamlfmt-*-"+filepath.Base(path))
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.Write(formatted)
	if err = errors.Join(err, tmp.Close()); err != nil {
		return err
	}
	color := "never"
	if useColor() {
		color = "always"
	}
	var stderr bytes.Buffer
	cmd := exec.CommandContext(
		ctx,
		"git",
		"--no-pager",
		"diff",
		"--no-index",
		"--color="+color,
		"--",
		path,
		tmp.Name(),
	)
	cmd.Stdout = w
	cmd.Stderr = &stderr
	err = cmd.Run()
	// git diff exits with status 1 when the files differ.
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok && exitErr.ExitCode() == 1 {
		return nil
	}
	if err != nil && stderr.Len() > 0 {
		return fmt.Errorf("git diff: %w: %s", err, bytes.TrimSpace(stderr.Bytes()))
	}
	return err
}

// notifyContext returns a context that is cancelled with a *signalError when
// the process receives SIGINT or SIGTERM. Unlike signal.NotifyContext, the
// cause keeps the signal itself, so main can exit with the matching status. A
// second signal gets the default behaviour and stops the process at once.
func notifyContext(parent context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancelCause(parent)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case s := <-sigs:
			signal.Stop(sigs)
			sig, _ := s.(syscall.Signal)
			cancel(&signalError{sig: sig})
		case <-ctx.Done():
		}
	}()
	return ctx, func() {
		signal.Stop(sigs)
		cancel(nil)
	}
}
