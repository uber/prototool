// Copyright (c) 2018 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package location

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	var base Path
	t.Run("Scope", func(t *testing.T) {
		foo := base.Scope(Message, 0)
		bar := base.Scope(Message, 0)
		assert.Equal(t, foo, bar)

		// Updating one should not change the other.
		foo = foo.Scope(Field, 0)
		assert.Equal(t, Path{4, 0, 2, 0}, foo)
		assert.NotEqual(t, foo, bar)
	})
	t.Run("Target", func(t *testing.T) {
		foo := base.Target(Name)
		bar := base.Target(Name)
		assert.Equal(t, foo, bar)

		// Updating one should not change the other.
		foo = foo.Target(Name)
		assert.Equal(t, Path{1, 1}, foo)
		assert.NotEqual(t, foo, bar)
	})
}
