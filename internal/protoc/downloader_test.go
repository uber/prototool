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

package protoc

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/prototool/internal/settings"
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

func TestNewDownloaderProtocValidation(t *testing.T) {
	tests := []struct {
		desc          string
		url           string
		binPath       string
		wktPath       string
		createBinPath bool
		createWKTPath bool
		err           error
	}{
		{
			desc: "No options",
		},
		{
			desc: "protocURL option",
			url:  "http://example.com",
		},
		{
			desc:          "protocBinPath with protocWKTPath",
			binPath:       "protoc",
			wktPath:       "include",
			createBinPath: true,
			createWKTPath: true,
		},
		{
			desc:    "protocURL set with protocBinPath and protocWKTPath",
			url:     "http://example.com",
			binPath: "protoc",
			wktPath: "include",
			err:     fmt.Errorf("cannot use protoc-url in combination with either protoc-bin-path or protoc-wkt-path"),
		},
		{
			desc:    "protocBinPath set without protocWKTPath",
			binPath: "protoc",
			err:     fmt.Errorf("both protoc-bin-path and protoc-wkt-path must be set"),
		},
		{
			desc:    "protocBinPath does not exist",
			binPath: "protoc",
			wktPath: "include",
			err:     fmt.Errorf("stat protoc: no such file or directory"),
		},
		{
			desc:          "protocWKTPath does not exist",
			binPath:       "protoc",
			wktPath:       "include",
			createBinPath: true,
			err:           fmt.Errorf("stat include: no such file or directory"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tmpRoot, err := ioutil.TempDir("", "test")
			require.NoError(t, err)

			// Clean up all the created test directories.
			defer func() {
				_ = os.RemoveAll(tmpRoot)
			}()

			if tt.createBinPath {
				tt.binPath, err = ioutil.TempDir(tmpRoot, tt.binPath)
				require.NoError(t, err)
			}

			if tt.createWKTPath {
				tt.wktPath, err = ioutil.TempDir(tmpRoot, tt.wktPath)
				require.NoError(t, err)
			}

			if tt.createBinPath && tt.createWKTPath {
				require.NoError(t, os.MkdirAll(filepath.Join(tt.wktPath, "google", "protobuf"), 0755))
			}

			_, err = newDownloader(
				settings.Config{},
				DownloaderWithProtocURL(tt.url),
				DownloaderWithProtocBinPath(tt.binPath),
				DownloaderWithProtocWKTPath(tt.wktPath),
			)

			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestNewDownloaderProtocURL(t *testing.T) {
	tests := []struct {
		desc            string
		expectedURL     string
		expectError     bool
		goarch          string
		goos            string
		protobufVersion string
	}{
		{
			protobufVersion: "3.10.0",
			goos:            "linux",
			goarch:          "amd64",
			expectedURL:     "https://github.com/protocolbuffers/protobuf/releases/download/v3.10.0/protoc-3.10.0-linux-x86_64.zip",
			desc:            "linux amd64 official version",
		},
		{
			protobufVersion: "3.10.0-rc-1",
			goos:            "linux",
			goarch:          "amd64",
			expectedURL:     "https://github.com/protocolbuffers/protobuf/releases/download/v3.10.0-rc1/protoc-3.10.0-rc-1-linux-x86_64.zip",
			desc:            "linux amd64 release candidate",
		},
		{
			protobufVersion: "3.10.0",
			goos:            "darwin",
			goarch:          "amd64",
			expectedURL:     "https://github.com/protocolbuffers/protobuf/releases/download/v3.10.0/protoc-3.10.0-osx-x86_64.zip",
			desc:            "darwin amd64 official version",
		},
		{
			protobufVersion: "3.10.0-rc-1",
			goos:            "darwin",
			goarch:          "amd64",
			expectedURL:     "https://github.com/protocolbuffers/protobuf/releases/download/v3.10.0-rc1/protoc-3.10.0-rc-1-osx-x86_64.zip",
			desc:            "darwin amd64 release candidate",
		},
		{
			goos:        "foo",
			goarch:      "amd64",
			expectError: true,
			desc:        "invalid goos",
		},
		{
			goos:        "linux",
			goarch:      "foo",
			expectError: true,
			desc:        "invalid goarch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			dl, err := newDownloader(
				settings.Config{
					Compile: settings.CompileConfig{
						ProtobufVersion: tt.protobufVersion,
					},
				},
			)

			url, err := dl.getProtocURL(tt.goos, tt.goarch)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, url, tt.expectedURL)
			}
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
