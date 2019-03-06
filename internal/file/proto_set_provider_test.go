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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/prototool/internal/settings"
	"go.uber.org/zap"
)

// We need to use valid as a representation of "cwd" so we verify
// that we do the recursive search properly. We used to use actual cwd,
// however since we added testdata/invalid, this will not work anymore.
// This is why we have the subdirectory "base" inside valid.
func TestProtoSetProviderGetMultipleForDirCwdAsValidRel(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	validDirPath := filepath.Join(cwd, "testdata", "valid")
	protoSetProvider := newTestProtoSetProvider(t)
	protoSets, err := protoSetProvider.getMultipleForDir(cwd, filepath.Join(".", "testdata", "valid"))
	require.NoError(t, err)
	require.Equal(
		t,
		[]*ProtoSet{
			&ProtoSet{
				WorkDirPath: cwd,
				DirPath:     validDirPath,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/valid/base/a": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/a/file.proto",
							DisplayPath: "testdata/valid/base/a/file.proto",
						},
					},
					cwd + "/testdata/valid/base/c": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/c/file.proto",
							DisplayPath: "testdata/valid/base/c/file.proto",
						},
					},
					cwd + "/testdata/valid/base/a/e": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/a/e/file.proto",
							DisplayPath: "testdata/valid/base/a/e/file.proto",
						},
					},
					cwd + "/testdata/valid/base/a/f": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/a/f/file.proto",
							DisplayPath: "testdata/valid/base/a/f/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/valid/base",
					ExcludePrefixes: []string{
						cwd + "/testdata/valid/base/c/i",
						cwd + "/testdata/valid/base/d",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion:       "3.4.0",
						IncludePaths:          []string{},
						IncludeWellKnownTypes: true,
					},
					Lint: settings.LintConfig{
						IncludeIDs:          []string{},
						ExcludeIDs:          []string{},
						IgnoreIDToFilePaths: map[string][]string{},
					},
					Gen: settings.GenConfig{
						GoPluginOptions: settings.GenGoPluginOptions{},
						Plugins:         []settings.GenPlugin{},
					},
				},
			},
			&ProtoSet{
				WorkDirPath: cwd,
				DirPath:     validDirPath,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/valid/base/a/d": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/a/d/file.proto",
							DisplayPath: "testdata/valid/base/a/d/file.proto",
						},
						{
							Path:        cwd + "/testdata/valid/base/a/d/file2.proto",
							DisplayPath: "testdata/valid/base/a/d/file2.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/valid/base/a/d",
					ExcludePrefixes: []string{
						cwd + "/testdata/valid/base/a/d/file3.proto",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion:       "3.2.0",
						IncludePaths:          []string{},
						IncludeWellKnownTypes: true,
					},
					Lint: settings.LintConfig{
						IncludeIDs:          []string{},
						ExcludeIDs:          []string{},
						IgnoreIDToFilePaths: map[string][]string{},
					},
					Gen: settings.GenConfig{
						GoPluginOptions: settings.GenGoPluginOptions{},
						Plugins:         []settings.GenPlugin{},
					},
				},
			},
			&ProtoSet{
				WorkDirPath: cwd,
				DirPath:     validDirPath,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/valid/base/b": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/b/file.proto",
							DisplayPath: "testdata/valid/base/b/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/valid/base/b",
					ExcludePrefixes: []string{
						cwd + "/testdata/valid/base/b/g/h",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion:       "3.3.0",
						IncludePaths:          []string{},
						IncludeWellKnownTypes: true,
					},
					Lint: settings.LintConfig{
						IncludeIDs:          []string{},
						ExcludeIDs:          []string{},
						IgnoreIDToFilePaths: map[string][]string{},
					},
					Gen: settings.GenConfig{
						GoPluginOptions: settings.GenGoPluginOptions{},
						Plugins:         []settings.GenPlugin{},
					},
				},
			},
		},
		protoSets,
	)
}

// We need to use valid as a representation of "cwd" so we verify
// that we do the recursive search properly. We used to use actual cwd,
// however since we added testdata/invalid, this will not work anymore.
// This is why we have the subdirectory "base" inside valid.
func TestProtoSetProviderGetMultipleForDirCwdAbs(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	validDirPath := filepath.Join(cwd, "testdata", "valid")
	protoSetProvider := newTestProtoSetProvider(t)
	protoSets, err := protoSetProvider.getMultipleForDir(cwd, validDirPath)
	require.NoError(t, err)
	require.Equal(
		t,
		[]*ProtoSet{
			&ProtoSet{
				WorkDirPath: cwd,
				DirPath:     validDirPath,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/valid/base/a": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/a/file.proto",
							DisplayPath: "testdata/valid/base/a/file.proto",
						},
					},
					cwd + "/testdata/valid/base/c": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/c/file.proto",
							DisplayPath: "testdata/valid/base/c/file.proto",
						},
					},
					cwd + "/testdata/valid/base/a/e": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/a/e/file.proto",
							DisplayPath: "testdata/valid/base/a/e/file.proto",
						},
					},
					cwd + "/testdata/valid/base/a/f": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/a/f/file.proto",
							DisplayPath: "testdata/valid/base/a/f/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/valid/base",
					ExcludePrefixes: []string{
						cwd + "/testdata/valid/base/c/i",
						cwd + "/testdata/valid/base/d",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion:       "3.4.0",
						IncludePaths:          []string{},
						IncludeWellKnownTypes: true,
					},
					Lint: settings.LintConfig{
						IncludeIDs:          []string{},
						ExcludeIDs:          []string{},
						IgnoreIDToFilePaths: map[string][]string{},
					},
					Gen: settings.GenConfig{
						GoPluginOptions: settings.GenGoPluginOptions{},
						Plugins:         []settings.GenPlugin{},
					},
				},
			},
			&ProtoSet{
				WorkDirPath: cwd,
				DirPath:     validDirPath,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/valid/base/a/d": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/a/d/file.proto",
							DisplayPath: "testdata/valid/base/a/d/file.proto",
						},
						{
							Path:        cwd + "/testdata/valid/base/a/d/file2.proto",
							DisplayPath: "testdata/valid/base/a/d/file2.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/valid/base/a/d",
					ExcludePrefixes: []string{
						cwd + "/testdata/valid/base/a/d/file3.proto",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion:       "3.2.0",
						IncludePaths:          []string{},
						IncludeWellKnownTypes: true,
					},
					Lint: settings.LintConfig{
						IncludeIDs:          []string{},
						ExcludeIDs:          []string{},
						IgnoreIDToFilePaths: map[string][]string{},
					},
					Gen: settings.GenConfig{
						GoPluginOptions: settings.GenGoPluginOptions{},
						Plugins:         []settings.GenPlugin{},
					},
				},
			},
			&ProtoSet{
				WorkDirPath: cwd,
				DirPath:     validDirPath,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/valid/base/b": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/b/file.proto",
							DisplayPath: "testdata/valid/base/b/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/valid/base/b",
					ExcludePrefixes: []string{
						cwd + "/testdata/valid/base/b/g/h",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion:       "3.3.0",
						IncludePaths:          []string{},
						IncludeWellKnownTypes: true,
					},
					Lint: settings.LintConfig{
						IncludeIDs:          []string{},
						ExcludeIDs:          []string{},
						IgnoreIDToFilePaths: map[string][]string{},
					},
					Gen: settings.GenConfig{
						GoPluginOptions: settings.GenGoPluginOptions{},
						Plugins:         []settings.GenPlugin{},
					},
				},
			},
		},
		protoSets,
	)
}

func TestProtoSetProviderGetMultipleForDirCwdSubRel(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	protoSetProvider := newTestProtoSetProvider(t)
	protoSets, err := protoSetProvider.getMultipleForDir(cwd, "testdata/valid/base/d/g")
	require.NoError(t, err)
	require.Equal(
		t,
		[]*ProtoSet{
			&ProtoSet{
				WorkDirPath: cwd,
				DirPath:     cwd + "/testdata/valid/base/d/g",
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/valid/base/d": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/d/file.proto",
							DisplayPath: "testdata/valid/base/d/file.proto",
						},
					},
					cwd + "/testdata/valid/base/d/g/h": []*ProtoFile{
						{
							Path:        cwd + "/testdata/valid/base/d/g/h/file.proto",
							DisplayPath: "testdata/valid/base/d/g/h/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath:         cwd + "/testdata/valid/base/d",
					ExcludePrefixes: []string{},
					Compile: settings.CompileConfig{
						ProtobufVersion:       "3.3.0",
						IncludePaths:          []string{},
						IncludeWellKnownTypes: true,
					},
					Lint: settings.LintConfig{
						IncludeIDs:          []string{},
						ExcludeIDs:          []string{},
						IgnoreIDToFilePaths: map[string][]string{},
					},
					Gen: settings.GenConfig{
						GoPluginOptions: settings.GenGoPluginOptions{},
						Plugins:         []settings.GenPlugin{},
					},
				},
			},
		},
		protoSets,
	)
}

func TestProtoSetProviderGetMultipleForDirTwoConfigFiles(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	protoSetProvider := newTestProtoSetProvider(t)
	_, err = protoSetProvider.getMultipleForDir(cwd, "testdata/invalid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "multiple configuration files")
}

func TestProtoSetProviderGetForDirEmpty(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	protoSetProvider := newTestProtoSetProvider(t)
	protoSet, err := protoSetProvider.GetForDir(dir, dir)
	assert.NoError(t, err)
	assert.Equal(
		t,
		&ProtoSet{
			WorkDirPath:    dir,
			DirPath:        dir,
			DirPathToFiles: map[string][]*ProtoFile{},
			Config: settings.Config{
				DirPath:         dir,
				ExcludePrefixes: []string{},
				Compile: settings.CompileConfig{
					IncludePaths:          []string{},
					IncludeWellKnownTypes: true,
				},
				Lint: settings.LintConfig{
					IncludeIDs:          []string{},
					ExcludeIDs:          []string{},
					IgnoreIDToFilePaths: map[string][]string{},
				},
				Gen: settings.GenConfig{
					GoPluginOptions: settings.GenGoPluginOptions{},
					Plugins:         []settings.GenPlugin{},
				},
			},
		},
		protoSet,
	)
}

func TestProtoSetProviderGetForDirConfigData(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	protoSetProvider := newTestProtoSetProviderConfigData(t, `{"protoc":{"version":"3.5.1"}}`)
	protoSet, err := protoSetProvider.GetForDir(dir, dir)
	assert.NoError(t, err)
	assert.Equal(
		t,
		&ProtoSet{
			WorkDirPath:    dir,
			DirPath:        dir,
			DirPathToFiles: map[string][]*ProtoFile{},
			Config: settings.Config{
				DirPath:         dir,
				ExcludePrefixes: []string{},
				Compile: settings.CompileConfig{
					ProtobufVersion:       "3.5.1",
					IncludePaths:          []string{},
					IncludeWellKnownTypes: true,
				},
				Lint: settings.LintConfig{
					IncludeIDs:          []string{},
					ExcludeIDs:          []string{},
					IgnoreIDToFilePaths: map[string][]string{},
				},
				Gen: settings.GenConfig{
					GoPluginOptions: settings.GenGoPluginOptions{},
					Plugins:         []settings.GenPlugin{},
				},
			},
		},
		protoSet,
	)
}

func newTestProtoSetProvider(t *testing.T) *protoSetProvider {
	return newProtoSetProvider(ProtoSetProviderWithLogger(newTestLogger(t)))
}

func newTestProtoSetProviderConfigData(t *testing.T, configData string) *protoSetProvider {
	return newProtoSetProvider(ProtoSetProviderWithLogger(newTestLogger(t)), ProtoSetProviderWithConfigData(configData))
}

func newTestLogger(t *testing.T) *zap.Logger {
	return zap.NewNop()
}
