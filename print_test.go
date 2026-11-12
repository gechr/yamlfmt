package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintFilesSingleInputHasNoHeader(t *testing.T) {
	var out strings.Builder
	err := printFiles(&out, strings.NewReader("a:   1\n"), nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "a: 1\n", out.String())
}

func TestPrintFilesMultipleInputs(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	// The first file has no final newline, which must not join it to the
	// next header.
	writeFile(t, "one.yaml", "x:  1")
	writeFile(t, "two.yaml", "y:   2\n")
	require.NoError(t, os.Symlink("two.yaml", "link.yaml"))

	var out strings.Builder
	err := printFiles(
		&out,
		strings.NewReader("z: 3\n"),
		[]string{"one.yaml", "-", "link.yaml"},
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, "# one.yaml\n---\nx: 1\n"+
		"# <stdin>\n---\nz: 3\n"+
		"# link.yaml -> two.yaml\n---\ny: 2\n", out.String())
}

func TestPrintFilesFramesEachInput(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	inputs := map[string]string{
		"empty.yaml":     "",
		"comment.yaml":   "# only\n",
		"marker.yaml":    "---\n---\na: 1\n",
		"versioned.yaml": "%YAML 1.1\n---\nb: 2\n",
		"plain.yaml":     "c: 3\n",
	}
	paths := []string{
		"versioned.yaml",
		"empty.yaml",
		"comment.yaml",
		"marker.yaml",
		"versioned.yaml",
		"plain.yaml",
	}
	for name, content := range inputs {
		writeFile(t, name, content)
	}

	var out strings.Builder
	require.NoError(t, printFiles(&out, nil, paths, nil))
	assert.Equal(t, "# versioned.yaml\n...\n%YAML 1.1\n---\nb: 2\n"+
		"# empty.yaml\n"+
		"# comment.yaml\n# only\n"+
		"# marker.yaml\n---\n---\na: 1\n"+
		"# versioned.yaml\n...\n%YAML 1.1\n---\nb: 2\n"+
		"# plain.yaml\n---\nc: 3\n", out.String())

	var want []content
	for _, path := range paths {
		want = append(want, loadStream(t, inputs[path])...)
	}
	assert.Equal(t, want, loadStream(t, out.String()))
}

func TestPrintFilesQuotesNamesThatWouldBreakTheHeader(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, "a\nb: injected", "x: 1\n")
	writeFile(t, "c.yaml", "y: 2\n")

	var out strings.Builder
	err := printFiles(&out, nil, []string{"a\nb: injected", "c.yaml"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "# \"a\\nb: injected\"\n---\nx: 1\n# c.yaml\n---\ny: 2\n", out.String())
}

func TestPrintFilesKeepsGoingAfterAnError(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, "bad.yaml", "a: 1\n---\nb: [\n")
	writeFile(t, "good.yaml", "c: 3\n")

	var out strings.Builder
	err := printFiles(&out, nil, []string{"missing.yaml", "bad.yaml", "good.yaml"}, nil)
	require.ErrorIs(t, err, os.ErrNotExist)
	lines := strings.Split(err.Error(), "\n")
	require.Len(t, lines, 2)
	assert.Equal(t, "missing.yaml: open missing.yaml: no such file or directory", lines[0])
	assert.True(t, strings.HasPrefix(lines[1], "bad.yaml: "), lines[1])
	assert.Equal(t, "# bad.yaml\n---\na: 1\n# good.yaml\n---\nc: 3\n", out.String())
}

func TestPrintFilesRejectsStdinTwice(t *testing.T) {
	err := printFiles(&strings.Builder{}, strings.NewReader(""), []string{"-", "-"}, nil)
	require.ErrorIs(t, err, errStdinTwice)
}

func TestPrintFilesReturnsWriteError(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, "a.yaml", "a: 1\n")
	writeFile(t, "b.yaml", "b: 2\n")

	err := printFiles(failingWriter{}, nil, []string{"a.yaml", "b.yaml"}, nil)
	require.ErrorIs(t, err, errWrite)

	err = printFiles(failingWriter{}, strings.NewReader("a: 1\n"), nil, nil)
	require.ErrorIs(t, err, errWrite)
}

func TestPrintFilesHighlights(t *testing.T) {
	var out strings.Builder
	err := printFiles(&out, strings.NewReader("a: 1\n"), nil, newHighlighter())
	require.NoError(t, err)
	assert.NotEqual(t, "a: 1\n", out.String())
	assert.True(
		t,
		strings.HasPrefix(out.String(), "\x1b["),
		"output is not highlighted: %q",
		out.String(),
	)
}

var errWrite = errors.New("write failed")

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errWrite
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Clean(path), []byte(content), 0o644))
}
