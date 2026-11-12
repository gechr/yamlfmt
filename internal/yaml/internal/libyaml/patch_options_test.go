// Copyright 2025 The go-yaml Project Contributors
// SPDX-License-Identifier: Apache-2.0

package libyaml

import (
	"sync"
	"testing"

	"github.com/gechr/yamlfmt/internal/yaml/internal/testutil/assert"
)

// TestWithLineWidthCanBeShared checks that one WithLineWidth option can be
// applied from several goroutines at once, like every other option. Run it
// with -race.
func TestWithLineWidthCanBeShared(t *testing.T) {
	opt := WithLineWidth(-5)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			var o Options
			assert.NoError(t, opt(&o))
			assert.Equal(t, -1, o.LineWidth)
		})
	}
	wg.Wait()
}
