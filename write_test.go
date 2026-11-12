package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewriteFilesWrite(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, "a.yaml", "a:   1\n")
	require.NoError(t, os.Chmod("a.yaml", 0o640))

	var out strings.Builder
	require.NoError(
		t,
		rewriteFiles(t.Context(), &out, options{write: true, paths: []string{"a.yaml"}}),
	)

	assert.Empty(t, out.String())
	assert.Equal(t, "a: 1\n", readFile(t, "a.yaml"))
	info, err := os.Stat("a.yaml")
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())
	assert.Equal(t, []string{"a.yaml"}, listDir(t, "."))
}

func TestRewriteFilesWriteThroughSymlink(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.Mkdir("real", 0o755))
	writeFile(t, "real/a.yaml", "a:   1\n")
	require.NoError(t, os.Symlink("real/a.yaml", "link.yaml"))

	// The file is listed three times, but formatted once.
	paths := []string{"link.yaml", "real/a.yaml", "./link.yaml"}
	require.NoError(
		t,
		rewriteFiles(t.Context(), &strings.Builder{}, options{write: true, paths: paths}),
	)

	info, err := os.Lstat("link.yaml")
	require.NoError(t, err)
	assert.Equal(t, os.ModeSymlink, info.Mode().Type(), "link.yaml was replaced by a regular file")
	assert.Equal(t, "a: 1\n", readFile(t, "real/a.yaml"))
	assert.Equal(t, []string{"a.yaml"}, listDir(t, "real"))
}

func TestRewriteFilesLeavesFormattedFileUntouched(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, "a.yaml", "a: 1\n")
	old := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes("a.yaml", old, old))

	require.NoError(
		t,
		rewriteFiles(
			t.Context(),
			&strings.Builder{},
			options{write: true, paths: []string{"a.yaml"}},
		),
	)

	info, err := os.Stat("a.yaml")
	require.NoError(t, err)
	assert.Equal(t, old, info.ModTime().UTC())
}

func TestRewriteFilesKeepsGoingAfterAnError(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, "bad.yaml", "a: [\n")
	writeFile(t, "good.yaml", "b:   2\n")
	require.NoError(t, os.Mkdir("dir.yaml", 0o755))

	err := rewriteFiles(t.Context(), &strings.Builder{}, options{
		write: true,
		paths: []string{"bad.yaml", "dir.yaml", "missing.yaml", "good.yaml"},
	})
	require.Error(t, err)
	lines := strings.Split(err.Error(), "\n")
	require.Len(t, lines, 3)
	assert.True(t, strings.HasPrefix(lines[0], "bad.yaml: "), lines[0])
	assert.Equal(t, "dir.yaml: not a regular file", lines[1])
	assert.True(t, strings.HasPrefix(lines[2], "missing.yaml: "), lines[2])
	require.ErrorIs(t, err, errNotRegular)

	assert.Equal(t, "a: [\n", readFile(t, "bad.yaml"))
	assert.Equal(t, "b: 2\n", readFile(t, "good.yaml"))
	assert.Equal(t, []string{"bad.yaml", "dir.yaml", "good.yaml"}, listDir(t, "."))
}

func TestRewriteFilesRejectsStdin(t *testing.T) {
	for _, paths := range [][]string{nil, {"-"}, {"a.yaml", "-"}} {
		err := rewriteFiles(t.Context(), &strings.Builder{}, options{write: true, paths: paths})
		require.ErrorIs(t, err, errStdin)
	}
}

func TestRewriteFilesDiff(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	t.Setenv("NO_COLOR", "1")
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, "a.yaml", "a:   1\n")
	writeFile(t, "b.yaml", "b: 2\n")

	tests := []struct {
		name  string
		write bool
		want  string
	}{
		{name: "diff only", write: false, want: "a:   1\n"},
		{name: "diff and write", write: true, want: "a: 1\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			err := rewriteFiles(
				t.Context(),
				&out,
				options{diff: true, write: tt.write, paths: []string{"a.yaml", "b.yaml"}},
			)
			require.NoError(t, err)
			assert.Equal(t, []string{"-a:   1", "+a: 1"}, changedLines(out.String()))
			assert.Equal(t, tt.want, readFile(t, "a.yaml"))
		})
	}
}

func TestRewriteFilesStopsWhenCancelled(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, "a.yaml", "a:   1\n")
	sigErr := &signalError{sig: syscall.SIGINT}
	ctx, cancel := context.WithCancelCause(t.Context())
	cancel(sigErr)

	err := rewriteFiles(ctx, &strings.Builder{}, options{write: true, paths: []string{"a.yaml"}})
	require.ErrorIs(t, err, sigErr)
	assert.Equal(t, "a:   1\n", readFile(t, "a.yaml"))
}

func TestNotifyContextRecordsSignal(t *testing.T) {
	ctx, stop := notifyContext(t.Context())
	defer stop()
	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGTERM))
	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("context was not cancelled by SIGTERM")
	}
	sigErr, ok := errors.AsType[*signalError](context.Cause(ctx))
	require.True(t, ok, "cause is %v", context.Cause(ctx))
	assert.Equal(t, syscall.SIGTERM, sigErr.sig)
}

func TestExitCode(t *testing.T) {
	assert.Equal(t, 0, exitCode(nil))
	assert.Equal(t, 1, exitCode(errNotRegular))
	assert.Equal(t, 130, exitCode(fmt.Errorf("wrapped: %w", &signalError{sig: syscall.SIGINT})))
	assert.Equal(t, 143, exitCode(&signalError{sig: syscall.SIGTERM}))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Clean(path))
	require.NoError(t, err)
	return string(b)
}

func listDir(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// changedLines returns the removed and added lines of a unified diff.
func changedLines(diff string) []string {
	var lines []string
	for line := range strings.SplitSeq(diff, "\n") {
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			continue
		}
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "+") {
			lines = append(lines, line)
		}
	}
	return lines
}

func TestResolveJobsDropsPathsToTheSameFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, "a.yaml", "a: 1\n")
	writeFile(t, "b.yaml", "b: 2\n")
	require.NoError(t, os.Symlink("a.yaml", "link.yaml"))

	jobs := resolveJobs(
		[]string{"a.yaml", "link.yaml", "./a.yaml", "b.yaml", "missing.yaml", "missing.yaml"},
	)
	paths := make([]string, 0, len(jobs))
	for _, job := range jobs {
		paths = append(paths, job.path)
	}
	assert.Equal(t, []string{"a.yaml", "b.yaml", "missing.yaml"}, paths)
}
