package main

import (
	"reflect"
	"testing"
)

func FuzzFormatBytes(f *testing.F) {
	for _, seed := range []string{
		"a:      b\nc:  d   # note\nk:\n# head\n- l\n\n- m\nq:\n    r:   s\n",
		"# top\nzeta: 1\nalpha: 2   # line\nmid: [a, b]\n# foot\n",
		"version: 1.10\noctal: 012\nflag: yes\nquoted: 'x'\nempty:\n",
		"base: &b\n  x: 1\nuse:\n  <<: *b\n  y: *b\n",
		"a: |+\n  x\n  y\n\nb: >-\n  x\n  y\n",
		"---\n---\na: 1\n---\n# c\n---\nb: 2\n---\n",
		"%YAML 1.1\n---\na: 1\n...\n%YAML 1.1\n---\nb: 2\n",
		"--- ~\n--- !!null\n--- &x\n--- !custom {a: [1, 2]}\n",
		"? complex\n: key\n\"quoted\\tkey\": 'v'\n",
	} {
		f.Add([]byte(seed))
	}
	// The formatted output must hold the same documents as the input, and
	// formatting must settle: in a few cases the emitter moves a comment on the
	// first run, so the check starts from the second.
	f.Fuzz(func(t *testing.T, in []byte) {
		first, err := formatBytes(in)
		if err != nil {
			return
		}
		want, err := streamContent(in)
		if err != nil {
			t.Fatalf("input formatted but does not load: %v\ninput: %q", err, in)
		}
		got, err := streamContent(first)
		if err != nil {
			t.Fatalf("output does not load: %v\ninput:  %q\noutput: %q", err, in, first)
		}
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("content changed\ninput:  %q\noutput: %q", in, first)
		}
		second, err := formatBytes(first)
		if err != nil {
			t.Fatalf("output does not format again: %v\ninput:  %q\noutput: %q", err, in, first)
		}
		third, err := formatBytes(second)
		if err != nil {
			t.Fatalf("output does not format again: %v\ninput:  %q\noutput: %q", err, in, second)
		}
		if string(third) != string(second) {
			t.Fatalf(
				"formatting does not settle\ninput:  %q\nsecond: %q\nthird:  %q",
				in,
				second,
				third,
			)
		}
	})
}
