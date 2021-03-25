// Copyright (c) 2021 Uber Technologies, Inc.
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

// Primarily adapted from https://github.com/golang/go/blob/e86168430f0aab8f971763e4b00c2aae7bec55f0/src/cmd/gofmt/gofmt.go.
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDo(t *testing.T) {
	testDo(t, "", "", nil)
	testDo(t, "abc", "abc", nil)
	testDo(t, "abc", "abd", nil, "-abc", "+abd")
	testDo(t, "abc\nabc", "abc\nabd", nil, "-abc", "+abd")
	testDo(t, "abc\nabc", "abc\nabc", nil)
}

func testDo(
	t *testing.T,
	input string,
	output string,
	expectedError error,
	expectedDiffs ...string,
) {
	diff, err := Do([]byte(input), []byte(output), "")
	for _, expectedDiff := range expectedDiffs {
		assert.Contains(t, string(diff), expectedDiff, "diff does not contain %s", expectedDiff)
	}
	assert.Equal(t, expectedError, err)
}
