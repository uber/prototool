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

package protoc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultBasePath(t *testing.T) {
	tests := []struct {
		goos             string
		goarch           string
		xdgCacheHome     string
		home             string
		expectedBasePath string
		expectError      bool
	}{
		{
			goos:             "darwin",
			goarch:           "amd64",
			xdgCacheHome:     "/foo",
			home:             "/Users/alice",
			expectedBasePath: "/foo/prototool/Darwin/x86_64",
		},
		{
			goos:             "darwin",
			goarch:           "amd64",
			home:             "/Users/alice",
			expectedBasePath: "/Users/alice/Library/Caches/prototool/Darwin/x86_64",
		},
		{
			goos:             "linux",
			goarch:           "amd64",
			xdgCacheHome:     "/foo",
			home:             "/home/alice",
			expectedBasePath: "/foo/prototool/Linux/x86_64",
		},
		{
			goos:             "linux",
			goarch:           "amd64",
			home:             "/home/alice",
			expectedBasePath: "/home/alice/.cache/prototool/Linux/x86_64",
		},
		{
			goos:         "foo",
			goarch:       "amd64",
			xdgCacheHome: "/foo",
			home:         "/home/alice",
			expectError:  true,
		},
		{
			goos:         "linux",
			goarch:       "foo",
			xdgCacheHome: "/foo",
			home:         "/home/alice",
			expectError:  true,
		},
		{
			goos:        "linux",
			goarch:      "amd64",
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(strings.Join([]string{tt.goos, tt.goarch, tt.xdgCacheHome, tt.home}, " "), func(t *testing.T) {
			basePath, err := getDefaultBasePathInternal(tt.goos, tt.goarch, newTestGetenvFunc(tt.xdgCacheHome, tt.home))
			if tt.expectError {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.expectedBasePath, basePath)
		})
	}
}

func newTestGetenvFunc(xdgCacheHome string, home string) func(string) string {
	m := make(map[string]string)
	if xdgCacheHome != "" {
		m["XDG_CACHE_HOME"] = xdgCacheHome
	}
	if home != "" {
		m["HOME"] = home
	}
	return func(key string) string { return m[key] }
}
