// Copyright 2025 The go-yaml Project Contributors
// SPDX-License-Identifier: Apache-2.0

package libyaml

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/gechr/yamlfmt/internal/yaml/internal/testutil/assert"
)

// describeTree prints every field of a node tree that dumping could change.
func describeTree(n *Node) string {
	var b strings.Builder
	var walk func(n *Node, depth int)
	walk = func(n *Node, depth int) {
		alias := ""
		if n.Alias != nil {
			alias = n.Alias.Anchor
		}
		fmt.Fprintf(&b, "%*s%v tag=%q style=%v value=%q anchor=%q alias=%q\n",
			depth*2, "", n.Kind, n.Tag, n.Style, n.Value, n.Anchor, alias)
		for _, c := range n.Content {
			walk(c, depth+1)
		}
	}
	walk(n, 0)
	return b.String()
}

const dumpInput = "a: 1\nb: '2'\nc: &x [true, \"3\", 4.5]\nd: *x\ne: !custom value\n"

func loadNode(t *testing.T) *Node {
	t.Helper()
	var n Node
	assert.NoError(t, Load([]byte(dumpInput), &n))
	return &n
}

type nodeMarshaler struct{ n *Node }

func (m nodeMarshaler) MarshalYAML() (any, error) { return m.n, nil }

// TestDumpLeavesInputNodesUnchanged checks that dumping does not change the
// nodes it is given, wherever they appear in the value being dumped.
func TestDumpLeavesInputNodesUnchanged(t *testing.T) {
	cases := map[string]func(n *Node) any{
		"document node":        func(n *Node) any { return n },
		"content node":         func(n *Node) any { return n.Content[0] },
		"node value in struct": func(n *Node) any { return struct{ N Node }{*n.Content[0]} },
		"node pointer in map":  func(n *Node) any { return map[string]*Node{"k": n.Content[0]} },
		"marshaler":            func(n *Node) any { return nodeMarshaler{n.Content[0]} },
	}
	for name, wrap := range cases {
		t.Run(name, func(t *testing.T) {
			n := loadNode(t)
			want := describeTree(n)
			_, err := Dump(wrap(n))
			assert.NoError(t, err)
			assert.Equal(t, want, describeTree(n))
		})
	}
}

// TestDumpSameNodeConcurrently checks that one node can be dumped from several
// goroutines at once. Run it with -race.
func TestDumpSameNodeConcurrently(t *testing.T) {
	n := loadNode(t)
	want, err := Dump(loadNode(t))
	assert.NoError(t, err)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			got, err := Dump(n)
			assert.NoError(t, err)
			assert.Equal(t, string(want), string(got))
		})
	}
	wg.Wait()
}

// TestDumpCyclicNode checks that a node graph whose alias refers back to its
// own anchor can be dumped.
func TestDumpCyclicNode(t *testing.T) {
	n := &Node{Kind: SequenceNode, Anchor: "a"}
	n.Content = []*Node{{Kind: AliasNode, Value: "a", Alias: n}}
	got, err := Dump(n)
	assert.NoError(t, err)
	assert.Equal(t, "&a\n- *a\n", string(got))
}
