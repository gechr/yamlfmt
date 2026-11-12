package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	yaml "github.com/gechr/yamlfmt/internal/yaml"
)

const indent = 2

// errContentChanged reports a document that the YAML library would not write
// back with the same content. The input is left as it was.
var errContentChanged = errors.New("formatting would change its content, so it was left unchanged")

// formatOptions controls how format frames documents.
type formatOptions struct {
	// continued marks the input as one of several written to a single stream,
	// so documents from other inputs may come before it.
	continued bool
}

// dumpOptions are the layout options for formatted output.
var dumpOptions = []yaml.Option{
	yaml.WithIndent(indent),
	yaml.WithCompactSeqIndent(false),
	yaml.WithLineWidth(-1),
}

// format reads a YAML stream from r and passes each formatted document to
// emit. The loader returns a document once it sees the next one's content, so
// output streams one document behind the input. emit must not keep the slice.
func format(r io.Reader, opts formatOptions, emit func([]byte) error) error {
	loader, err := yaml.NewLoader(r, yaml.WithStreamNodes())
	if err != nil {
		return err
	}
	var (
		docs   int
		stream *yaml.Stream
	)
	for {
		var node yaml.Node
		err := loader.Load(&node)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		var out []byte
		switch node.Kind {
		case yaml.StreamNode:
			// A stream node comes before each document and holds its
			// directives. The last one holds comments outside any document.
			stream = node.Stream
			out = appendComment(out, node.HeadComment)
			out = appendComment(out, node.FootComment)
		case yaml.DocumentNode:
			docs++
			out, err = appendDocument(out, &node, stream, docs > 1 || opts.continued)
			if err != nil {
				return fmt.Errorf("document %d: %w", docs, err)
			}
		case yaml.SequenceNode, yaml.MappingNode, yaml.ScalarNode, yaml.AliasNode:
			// The loader only returns these inside a document.
		}
		if len(out) > 0 {
			if err := emit(out); err != nil {
				return err
			}
		}
	}
}

// formatBytes formats a complete YAML stream held in memory.
func formatBytes(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	err := format(bytes.NewReader(src), formatOptions{}, func(b []byte) error {
		_, err := buf.Write(b)
		return err
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// appendDocument appends the formatted document doc. preceded reports that
// another document may come before it in the output. The "---" marker is
// written whenever leaving it out would change the stream: after another
// document, after a %YAML directive, and for an empty document, which would
// otherwise vanish. A directive after another document needs a "..." first.
func appendDocument(
	out []byte,
	doc *yaml.Node,
	stream *yaml.Stream,
	preceded bool,
) ([]byte, error) {
	versioned := stream != nil && stream.Version != nil
	if versioned {
		if preceded {
			out = append(out, "...\n"...)
		}
		out = fmt.Appendf(out, "%%YAML %d.%d\n", stream.Version.Major, stream.Version.Minor)
	}
	empty := isEmptyDocument(doc)
	if preceded || versioned || empty {
		out = append(out, "---\n"...)
	}
	if empty {
		// Dumping an empty document writes blank lines around its comments.
		// An end marker after them keeps the loader from moving them to the
		// previous document when this one is last.
		n := len(out)
		out = appendComment(out, doc.HeadComment)
		if len(doc.Content) == 1 {
			out = appendComment(out, doc.Content[0].HeadComment)
			out = appendComment(out, doc.Content[0].LineComment)
			out = appendComment(out, doc.Content[0].FootComment)
		}
		out = appendComment(out, doc.FootComment)
		if len(out) > n {
			out = append(out, "...\n"...)
		}
		return out, nil
	}
	want := contentOf(doc)
	b, err := yaml.Dump(doc, dumpOptions...)
	if err != nil {
		return nil, err
	}
	if err := checkContent(want, b); err != nil {
		return nil, err
	}
	return append(out, b...), nil
}

// content is the data held by a node and its children, without layout, quoting
// style or comments.
type content struct {
	kind     yaml.Kind
	tag      string
	value    string
	anchor   string
	alias    string
	children []content
}

func contentOf(n *yaml.Node) content {
	c := content{kind: n.Kind, tag: n.ShortTag(), value: n.Value, anchor: n.Anchor}
	if n.Alias != nil {
		c.alias = n.Alias.Anchor
	}
	if len(n.Content) > 0 {
		c.children = make([]content, len(n.Content))
		for i, child := range n.Content {
			c.children[i] = contentOf(child)
		}
	}
	return c
}

// checkContent reloads formatted and checks that it holds want, so that a bug
// in the emitter cannot change the data in a file.
func checkContent(want content, formatted []byte) error {
	var got yaml.Node
	if err := yaml.Load(formatted, &got); err != nil {
		return fmt.Errorf("%w: %w", errContentChanged, err)
	}
	if !reflect.DeepEqual(want, contentOf(&got)) {
		return errContentChanged
	}
	return nil
}

// isEmptyDocument reports whether doc has no content at all: not even an
// explicit null such as "~", "null" or "!!null", or an anchor.
func isEmptyDocument(doc *yaml.Node) bool {
	if len(doc.Content) == 0 {
		return true
	}
	if len(doc.Content) > 1 {
		return false
	}
	n := doc.Content[0]
	return n.Kind == yaml.ScalarNode && n.Tag == "!!null" && n.Value == "" &&
		n.Style == 0 && n.Anchor == ""
}

func appendComment(out []byte, comment string) []byte {
	if comment == "" {
		return out
	}
	out = append(out, comment...)
	if !strings.HasSuffix(comment, "\n") {
		out = append(out, '\n')
	}
	return out
}
