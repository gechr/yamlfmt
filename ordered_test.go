package main

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestForEachOrdered(t *testing.T) {
	const limit = 3
	items := make([]int, 50)
	for i := range items {
		items[i] = i
	}
	var running, maxRunning, pending, maxPending atomic.Int32
	var got []int
	forEachOrdered(items, limit, func(i int) int {
		n := running.Add(1)
		updateMax(&maxRunning, n)
		// Make early items slower than later ones.
		time.Sleep(time.Duration(len(items)-i) * 100 * time.Microsecond)
		running.Add(-1)
		updateMax(&maxPending, pending.Add(1))
		return i * 10
	}, func(i, r int) {
		pending.Add(-1)
		assert.Equal(t, i*10, r)
		got = append(got, i)
	})
	assert.Equal(t, items, got)
	assert.LessOrEqual(t, maxRunning.Load(), int32(limit))
	assert.LessOrEqual(t, maxPending.Load(), int32(limit))
}

func TestForEachOrderedEmpty(t *testing.T) {
	forEachOrdered(nil, 2, func(int) int {
		t.Fatal("work called for no items")
		return 0
	}, func(int, int) {
		t.Fatal("yield called for no items")
	})
}

func updateMax(m *atomic.Int32, n int32) {
	for {
		cur := m.Load()
		if n <= cur || m.CompareAndSwap(cur, n) {
			return
		}
	}
}
