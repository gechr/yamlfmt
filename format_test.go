package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	yaml "github.com/gechr/yamlfmt/internal/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "normalises spacing and indents sequences",
			in:   "a:      b\nc:  d   # note\nk:\n# head\n- l\n\n- m\nq:\n    r:   s\n    t       : u\n",
			want: "a: b\nc: d # note\nk:\n  # head\n  - l\n  - m\nq:\n  r: s\n  t: u\n",
		},
		{
			name: "keeps key order and comments",
			in:   "# top\nzeta: 1\nalpha: 2   # line\n# head\nmid: [a, b]\n# foot\n",
			want: "# top\nzeta: 1\nalpha: 2 # line\n# head\nmid: [a, b]\n# foot\n",
		},
		{
			name: "keeps scalars as written",
			in:   "version: 1.10\noctal: 012\nflag: yes\nquoted: 'x'\nempty:\n",
			want: "version: 1.10\noctal: 012\nflag: yes\nquoted: 'x'\nempty:\n",
		},
		{
			name: "keeps anchors, aliases and merge keys",
			in:   "base: &b\n  x: 1\nuse:\n  <<: *b\n  y: *b\n",
			want: "base: &b\n  x: 1\nuse:\n  <<: *b\n  y: *b\n",
		},
		{
			name: "keeps duplicate keys",
			in:   "a: 1\na: 2\n",
			want: "a: 1\na: 2\n",
		},
		{
			name: "keeps literal block scalars",
			in:   "a: |+\n  x\n  y\n\n",
			want: "a: |+\n  x\n  y\n\n",
		},
		{
			// The emitter re-flows folded scalars, and the line width is
			// unlimited so that it never wraps plain scalars.
			name: "re-flows folded block scalars onto one line",
			in:   "a: >-\n  x\n  y\n",
			want: "a: >-\n  x y\n",
		},
		{
			name: "ends a block scalar without a final newline on a line boundary",
			in:   "a: |-\n  x",
			want: "a: |-\n  x\n",
		},
		{
			name: "separates documents",
			in:   "# one\na: 1\n---\n# two\nb: 2 # c\n",
			want: "# one\na: 1\n---\n# two\nb: 2 # c\n",
		},
		{
			name: "drops a redundant leading marker",
			in:   "---\na: 1\n",
			want: "a: 1\n",
		},
		{
			name: "keeps a lone empty document",
			in:   "---\n",
			want: "---\n",
		},
		{
			name: "keeps leading, middle and trailing empty documents",
			in:   "---\n---\na: 1\n---\n---\nb: 2\n---\n",
			want: "---\n---\na: 1\n---\n---\nb: 2\n---\n",
		},
		{
			name: "keeps comments in an empty document",
			in:   "---\n# c\n---\na: 1\n",
			want: "---\n# c\n...\n---\na: 1\n",
		},
		{
			name: "keeps comments in a last empty document",
			in:   "a: 1\n---\n# c\n...\n",
			want: "a: 1\n---\n# c\n...\n",
		},
		{
			name: "keeps a comment after a null value",
			in:   "foo:\n# bar: 1\nbaz: 2\n",
			want: "foo:\n# bar: 1\nbaz: 2\n",
		},
		{
			name: "keeps explicit nulls",
			in:   "--- ~\n--- null\n--- !!null\n--- &x\n",
			want: "~\n---\nnull\n---\n!!null\n---\n&x\n",
		},
		{
			name: "keeps a comment-only stream",
			in:   "# only\n",
			want: "# only\n",
		},
		{
			name: "keeps the YAML version directive",
			in:   "%YAML 1.1\n---\na:   1\n",
			want: "%YAML 1.1\n---\na: 1\n",
		},
		{
			name: "keeps a YAML 1.2 directive",
			in:   "%YAML 1.2\n---\na:   1\n...\n%YAML 1.2\n---\nb: 2\n",
			want: "%YAML 1.2\n---\na: 1\n...\n%YAML 1.2\n---\nb: 2\n",
		},
		{
			name: "keeps a keep-chomped scalar before a comment",
			in:   "a: |+\n  x\n\n# c\n",
			want: "a: |+\n  x\n\n# c\n",
		},
		{
			name: "keeps a comment on an empty document's marker",
			in:   "--- # here\n...\n# end\n",
			want: "---\n# here\n# end\n...\n",
		},
		{
			name: "keeps a comment on a document's marker",
			in:   "--- # here\na: 1\n",
			want: "# here\na: 1\n",
		},
		{
			name: "ends the previous document before a directive",
			in:   "a: 1\n...\n%YAML 1.1\n---\nb: 2\n",
			want: "a: 1\n...\n%YAML 1.1\n---\nb: 2\n",
		},
		{
			name: "ends a previous scalar document before a directive",
			in:   "one\n...\n%YAML 1.1\n---\ntwo\n",
			want: "one\n...\n%YAML 1.1\n---\ntwo\n",
		},
		{
			name: "keeps several versioned documents",
			in:   "%YAML 1.1\n---\na: 1\n...\n%YAML 1.1\n---\nb: 2\n",
			want: "%YAML 1.1\n---\na: 1\n...\n%YAML 1.1\n---\nb: 2\n",
		},
		{
			name: "converts CRLF line endings to LF",
			in:   "a: 1\r\nb: 2\r\n",
			want: "a: 1\nb: 2\n",
		},
		{
			name: "formats empty input as empty output",
			in:   "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatBytes([]byte(tt.in))
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))

			again, err := formatBytes(got)
			require.NoError(t, err)
			assert.Equal(t, string(got), string(again), "formatting is not idempotent")
			assert.Equal(t, loadStream(t, tt.in), loadStream(t, string(got)))
		})
	}
}

func TestFormatBytesErrors(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "invalid syntax", in: "a: [\n"},
		{name: "invalid later document", in: "a: 1\n---\nb: [\n"},
		{name: "unsupported YAML version", in: "%YAML 2.0\n---\na: 1\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := formatBytes([]byte(tt.in))
			require.Error(t, err)
		})
	}
}

func TestCheckContent(t *testing.T) {
	var doc yaml.Node
	require.NoError(t, yaml.Load([]byte("a: 1\n"), &doc))
	want := contentOf(&doc)

	require.NoError(t, checkContent(want, []byte("a:   1\n")))
	require.ErrorIs(t, checkContent(want, []byte("a: 2\n")), errContentChanged)
	require.ErrorIs(t, checkContent(want, []byte("a: '1'\n")), errContentChanged)
	require.ErrorIs(t, checkContent(want, []byte("a: [\n")), errContentChanged)
}

func TestFormatContinued(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "marks the first document", in: "a: 1\n", want: "---\na: 1\n"},
		{
			name: "ends a previous input before a directive",
			in:   "%YAML 1.1\n---\na: 1\n",
			want: "...\n%YAML 1.1\n---\na: 1\n",
		},
		{name: "adds nothing for no documents", in: "# only\n", want: "# only\n"},
		{name: "keeps a leading empty document", in: "---\n---\na: 1\n", want: "---\n---\na: 1\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got strings.Builder
			err := format(
				strings.NewReader(tt.in),
				formatOptions{continued: true},
				func(b []byte) error {
					got.Write(b)
					return nil
				},
			)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestFormatEmitsDocumentsBeforeLaterError(t *testing.T) {
	var got []string
	err := format(
		bytes.NewReader([]byte("a:  1\n---\nb: [\n")),
		formatOptions{},
		func(b []byte) error {
			got = append(got, string(b))
			return nil
		},
	)
	require.Error(t, err)
	assert.Equal(t, []string{"a: 1\n"}, got)
}

func TestFormatReturnsEmitError(t *testing.T) {
	errEmit := errors.New("emit failed")
	err := format(bytes.NewReader([]byte("a: 1\n")), formatOptions{}, func([]byte) error {
		return errEmit
	})
	require.ErrorIs(t, err, errEmit)
}

func TestFormatStreams(t *testing.T) {
	r, w := io.Pipe()
	emitted := make(chan string, 2)
	done := make(chan error, 1)
	go func() {
		done <- format(r, formatOptions{}, func(b []byte) error {
			emitted <- string(b)
			return nil
		})
	}()

	// The loader returns a document once it sees the next one's content.
	_, err := w.Write([]byte("a: 1\n---\nb: 2\n"))
	require.NoError(t, err)
	select {
	case got := <-emitted:
		assert.Equal(t, "a: 1\n", got)
	case <-time.After(5 * time.Second):
		t.Fatal("first document was not emitted before the input ended")
	}

	require.NoError(t, w.Close())
	require.NoError(t, <-done)
	assert.Equal(t, "---\nb: 2\n", <-emitted)
}

// loadStream returns the content of each document in s.
func loadStream(t *testing.T, s string) []content {
	t.Helper()
	docs, err := streamContent([]byte(s))
	require.NoError(t, err)
	return docs
}

func streamContent(b []byte) ([]content, error) {
	loader, err := yaml.NewLoader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var docs []content
	for {
		var node yaml.Node
		err := loader.Load(&node)
		if errors.Is(err, io.EOF) {
			return docs, nil
		}
		if err != nil {
			return nil, err
		}
		docs = append(docs, contentOf(&node))
	}
}
