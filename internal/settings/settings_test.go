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

package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExternalConfigValidate(t *testing.T) {
	t.Run("gen.go_options.no_default_modifiers", func(t *testing.T) {
		e := ExternalConfig{}
		e.Gen.GoOptions.NoDefaultModifiers = true
		assert.EqualError(t, e.Validate(), "gen.go_options.no_default_modifiers is not a configurable setting")
	})
	t.Run("lint.group", func(t *testing.T) {
		e := ExternalConfig{}
		e.Lint.Group = "oh no"
		assert.EqualError(t, e.Validate(), "lint.group is not a configurable setting: have oh no")
	})
}
