// Copyright 2025 The go-yaml Project Contributors
// SPDX-License-Identifier: Apache-2.0

package libyaml

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/gechr/yamlfmt/internal/yaml/internal/testutil/assert"
)

// roundTrip loads every document in src as a node and dumps them again.
func roundTrip(t *testing.T, src string) string {
	t.Helper()
	loader, err := NewLoader(bytes.NewReader([]byte(src)))
	assert.NoError(t, err)
	var out bytes.Buffer
	dumper, err := NewDumper(&out, WithIndent(2))
	assert.NoError(t, err)
	for {
		var n Node
		err := loader.Load(&n)
		if errors.Is(err, io.EOF) {
			break
		}
		assert.NoError(t, err)
		assert.NoError(t, dumper.Dump(&n))
	}
	assert.NoError(t, dumper.Close())
	return out.String()
}

// decodeAll decodes every document in src.
func decodeAll(t *testing.T, src string) []any {
	t.Helper()
	loader, err := NewLoader(bytes.NewReader([]byte(src)))
	assert.NoError(t, err)
	var docs []any
	for {
		var v any
		err := loader.Load(&v)
		if errors.Is(err, io.EOF) {
			return docs
		}
		assert.NoError(t, err)
		docs = append(docs, v)
	}
}

// TestDumpKeepChompedScalarBeforeFootComment checks that a document ending
// in a keep-chomped block scalar and a comment keeps the scalar's value. A
// blank line written before the comment would become part of the scalar.
func TestDumpKeepChompedScalarBeforeFootComment(t *testing.T) {
	cases := map[string]string{
		"literal":              "a: |+\n  x\n\n# c\n",
		"folded":               "a: >+\n  x\n\n# c\n",
		"nested":               "a:\n  b: |+\n    x\n\n# c\n",
		"root":                 "--- |+\n  x\n\n# c\n",
		"no trailing blank":    "a: |+\n  x\n# c\n",
		"next document":        "a: |+\n  x\n\n# c\n---\nb: 1\n",
		"comment before value": "a: |+\n  x\n\n# c\nb: 1\n",
		"nested foot comment":  "a:\n  b: |+\n    x\n\n  # c\nd: 1\n",
		"last in sequence":     "- |+\n  x\n\n# c\n",
		"clipped":              "a: |\n  x\n\n# c\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			once := roundTrip(t, src)
			assert.DeepEqualf(t, decodeAll(t, src), decodeAll(t, once), "value changed: %q", once)
			assert.Equal(t, once, roundTrip(t, once))
		})
	}
}
