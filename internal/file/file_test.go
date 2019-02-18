// Copyright (c) 2019 Uber Technologies, Inc.
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

package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAbsClean(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	testAbsClean(t, "", "")
	testAbsClean(t, "/", "/")
	testAbsClean(t, "/path/to/foo", "/path/to/foo")
	testAbsClean(t, "/path/to/foo", "/path//to//foo")
	testAbsClean(t, filepath.Join(wd, "path/to/foo"), "path/to/foo")
	testAbsClean(t, filepath.Join(wd, "path/to/foo"), "path//to//foo")
}

func testAbsClean(t *testing.T, expected string, input string) {
	actual, err := AbsClean(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestCheckAbs(t *testing.T) {
	assert.NoError(t, CheckAbs("/path/to/foo"))
	assert.Error(t, CheckAbs("path/to/foo"))
}
