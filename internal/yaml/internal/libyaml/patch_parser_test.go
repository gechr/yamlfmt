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

// loadDocuments loads every document in src as a node.
func loadDocuments(t *testing.T, src string, opts ...Option) []*Node {
	t.Helper()
	loader, err := NewLoader(bytes.NewReader([]byte(src)), opts...)
	assert.NoError(t, err)
	var docs []*Node
	for {
		var n Node
		err := loader.Load(&n)
		if errors.Is(err, io.EOF) {
			return docs
		}
		assert.NoError(t, err)
		docs = append(docs, &n)
	}
}

// TestEmptyDocumentKeepsStartComment checks that a comment on the "---" line
// of a document with no content is kept on its empty scalar.
func TestEmptyDocumentKeepsStartComment(t *testing.T) {
	cases := []struct {
		name string
		src  string
		head string
		foot string
	}{
		{name: "end of stream", src: "--- # here\n", head: "# here"},
		{name: "explicit end", src: "--- # here\n...\n", head: "# here"},
		{name: "explicit end and footer", src: "--- # here\n...\n# end\n", head: "# here", foot: "# end"},
		{name: "next document", src: "--- # here\n---\na: 1\n", head: "# here"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := loadDocuments(t, tc.src)[0]
			assert.Equal(t, 1, len(doc.Content))
			assert.Equal(t, ScalarNode, doc.Content[0].Kind)
			assert.Equal(t, tc.head, doc.Content[0].HeadComment)
			assert.Equal(t, tc.foot, doc.FootComment)
		})
	}
}

// TestDocumentStartCommentBeforeContent checks that a comment on the "---"
// line of a document with content still goes to that content.
func TestDocumentStartCommentBeforeContent(t *testing.T) {
	doc := loadDocuments(t, "--- # here\na: 1\n")[0]
	assert.Equal(t, "# here", doc.Content[0].Content[0].HeadComment)
}

// TestVersionDirectives checks which %YAML versions load and dump.
func TestVersionDirectives(t *testing.T) {
	for _, version := range []string{"1.1", "1.2"} {
		t.Run(version, func(t *testing.T) {
			docs := loadDocuments(t, "%YAML "+version+"\n---\na: 1\n", WithStreamNodes())
			assert.Equal(t, StreamNode, docs[0].Kind)
			got := docs[0].Stream.Version
			assert.Equal(t, version, string(rune('0'+got.Major))+"."+string(rune('0'+got.Minor)))
		})
	}
	t.Run("2.0", func(t *testing.T) {
		loader, err := NewLoader(bytes.NewReader([]byte("%YAML 2.0\n---\na: 1\n")))
		assert.NoError(t, err)
		var n Node
		assert.ErrorMatches(t, ".*incompatible YAML document.*", loader.Load(&n))
	})
	t.Run("emit 1.2", func(t *testing.T) {
		events := []Event{
			NewStreamStartEvent(UTF8_ENCODING),
			NewDocumentStartEvent(NewVersionDirective(1, 2), nil, false),
			NewScalarEvent(nil, nil, []byte("a"), true, true, PLAIN_SCALAR_STYLE),
			NewDocumentEndEvent(true),
			NewStreamEndEvent(),
		}
		emitter := NewEmitter()
		var output []byte
		emitter.SetOutputString(&output)
		for i := range events {
			assert.NoError(t, emitter.Emit(&events[i]))
		}
		assert.Equal(t, "%YAML 1.2\n---\na\n...\n", string(output))
	})
}
