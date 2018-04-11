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

package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uber/prototool/internal/x/settings"
	"go.uber.org/zap"
)

func TestProtoSetProviderGetForFilesAll(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	protoSetProvider := newTestProtoSetProvider(t)
	protoSets, err := protoSetProvider.GetForFiles(
		cwd,
		"testdata/a/file.proto",
		"testdata/b/file.proto",
		"testdata/c/file.proto",
		"testdata/a/d/file.proto",
		"testdata/a/d/file2.proto",
		"testdata/a/d/file3.proto",
		"testdata/a/e/file.proto",
		"testdata/a/f/file.proto",
		"testdata/b/g/h/file.proto",
	)
	require.NoError(t, err)
	require.Equal(
		t,
		[]*ProtoSet{

			&ProtoSet{
				WorkDirPath: cwd,
				DirPath:     cwd,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/a": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/file.proto",
							DisplayPath: "testdata/a/file.proto",
						},
					},
					cwd + "/testdata/c": []*ProtoFile{
						{
							Path:        cwd + "/testdata/c/file.proto",
							DisplayPath: "testdata/c/file.proto",
						},
					},
					cwd + "/testdata/a/e": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/e/file.proto",
							DisplayPath: "testdata/a/e/file.proto",
						},
					},
					cwd + "/testdata/a/f": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/f/file.proto",
							DisplayPath: "testdata/a/f/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata",
					ExcludePrefixes: []string{
						cwd + "/testdata/c/i",
						cwd + "/testdata/d",
						cwd + "/testdata/vendor",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion: "3.4.0",
						IncludePaths:    []string{},
					},
					Lint: settings.LintConfig{
						IDs:                 []string{},
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
				DirPath:     cwd,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/a/d": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/d/file.proto",
							DisplayPath: "testdata/a/d/file.proto",
						},
						{
							Path:        cwd + "/testdata/a/d/file2.proto",
							DisplayPath: "testdata/a/d/file2.proto",
						},
						{
							Path:        cwd + "/testdata/a/d/file3.proto",
							DisplayPath: "testdata/a/d/file3.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/a/d",
					ExcludePrefixes: []string{
						cwd + "/testdata/a/d/file3.proto",
						cwd + "/testdata/a/d/vendor",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion: "3.2.0",
						IncludePaths:    []string{},
					},
					Lint: settings.LintConfig{
						IDs:                 []string{},
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
				DirPath:     cwd,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/b": []*ProtoFile{
						{
							Path:        cwd + "/testdata/b/file.proto",
							DisplayPath: "testdata/b/file.proto",
						},
					},
					cwd + "/testdata/b/g/h": []*ProtoFile{
						{
							Path:        cwd + "/testdata/b/g/h/file.proto",
							DisplayPath: "testdata/b/g/h/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/b",
					ExcludePrefixes: []string{
						cwd + "/testdata/b/g/h",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion: "3.3.0",
						IncludePaths:    []string{},
					},
					Lint: settings.LintConfig{
						IDs:                 []string{},
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

func TestProtoSetProviderGetForFilesSomeMissing(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	protoSetProvider := newTestProtoSetProvider(t)
	protoSets, err := protoSetProvider.GetForFiles(
		cwd,
		"testdata/a/file.proto",
		"testdata/c/file.proto",
		"testdata/a/d/file.proto",
		"testdata/a/d/file3.proto",
		"testdata/a/e/file.proto",
		"testdata/a/f/file.proto",
	)
	require.NoError(t, err)
	require.Equal(
		t,
		[]*ProtoSet{

			&ProtoSet{
				WorkDirPath: cwd,
				DirPath:     cwd,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/a": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/file.proto",
							DisplayPath: "testdata/a/file.proto",
						},
					},
					cwd + "/testdata/c": []*ProtoFile{
						{
							Path:        cwd + "/testdata/c/file.proto",
							DisplayPath: "testdata/c/file.proto",
						},
					},
					cwd + "/testdata/a/e": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/e/file.proto",
							DisplayPath: "testdata/a/e/file.proto",
						},
					},
					cwd + "/testdata/a/f": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/f/file.proto",
							DisplayPath: "testdata/a/f/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata",
					ExcludePrefixes: []string{
						cwd + "/testdata/c/i",
						cwd + "/testdata/d",
						cwd + "/testdata/vendor",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion: "3.4.0",
						IncludePaths:    []string{},
					},
					Lint: settings.LintConfig{
						IDs:                 []string{},
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
				DirPath:     cwd,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/a/d": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/d/file.proto",
							DisplayPath: "testdata/a/d/file.proto",
						},
						{
							Path:        cwd + "/testdata/a/d/file3.proto",
							DisplayPath: "testdata/a/d/file3.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/a/d",
					ExcludePrefixes: []string{
						cwd + "/testdata/a/d/file3.proto",
						cwd + "/testdata/a/d/vendor",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion: "3.2.0",
						IncludePaths:    []string{},
					},
					Lint: settings.LintConfig{
						IDs:                 []string{},
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

func TestProtoSetProviderGetForDirCwdRel(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	protoSetProvider := newTestProtoSetProvider(t)
	protoSets, err := protoSetProvider.GetForDir(cwd, ".")
	require.NoError(t, err)
	require.Equal(
		t,
		[]*ProtoSet{

			&ProtoSet{
				WorkDirPath: cwd,
				DirPath:     cwd,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/a": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/file.proto",
							DisplayPath: "testdata/a/file.proto",
						},
					},
					cwd + "/testdata/c": []*ProtoFile{
						{
							Path:        cwd + "/testdata/c/file.proto",
							DisplayPath: "testdata/c/file.proto",
						},
					},
					cwd + "/testdata/a/e": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/e/file.proto",
							DisplayPath: "testdata/a/e/file.proto",
						},
					},
					cwd + "/testdata/a/f": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/f/file.proto",
							DisplayPath: "testdata/a/f/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata",
					ExcludePrefixes: []string{
						cwd + "/testdata/c/i",
						cwd + "/testdata/d",
						cwd + "/testdata/vendor",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion: "3.4.0",
						IncludePaths:    []string{},
					},
					Lint: settings.LintConfig{
						IDs:                 []string{},
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
				DirPath:     cwd,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/a/d": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/d/file.proto",
							DisplayPath: "testdata/a/d/file.proto",
						},
						{
							Path:        cwd + "/testdata/a/d/file2.proto",
							DisplayPath: "testdata/a/d/file2.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/a/d",
					ExcludePrefixes: []string{
						cwd + "/testdata/a/d/file3.proto",
						cwd + "/testdata/a/d/vendor",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion: "3.2.0",
						IncludePaths:    []string{},
					},
					Lint: settings.LintConfig{
						IDs:                 []string{},
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
				DirPath:     cwd,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/b": []*ProtoFile{
						{
							Path:        cwd + "/testdata/b/file.proto",
							DisplayPath: "testdata/b/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/b",
					ExcludePrefixes: []string{
						cwd + "/testdata/b/g/h",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion: "3.3.0",
						IncludePaths:    []string{},
					},
					Lint: settings.LintConfig{
						IDs:                 []string{},
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

func TestProtoSetProviderGetForDirCwdAbs(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	protoSetProvider := newTestProtoSetProvider(t)
	protoSets, err := protoSetProvider.GetForDir(cwd, cwd)
	require.NoError(t, err)
	require.Equal(
		t,
		[]*ProtoSet{

			&ProtoSet{
				WorkDirPath: cwd,
				DirPath:     cwd,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/a": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/file.proto",
							DisplayPath: "testdata/a/file.proto",
						},
					},
					cwd + "/testdata/c": []*ProtoFile{
						{
							Path:        cwd + "/testdata/c/file.proto",
							DisplayPath: "testdata/c/file.proto",
						},
					},
					cwd + "/testdata/a/e": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/e/file.proto",
							DisplayPath: "testdata/a/e/file.proto",
						},
					},
					cwd + "/testdata/a/f": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/f/file.proto",
							DisplayPath: "testdata/a/f/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata",
					ExcludePrefixes: []string{
						cwd + "/testdata/c/i",
						cwd + "/testdata/d",
						cwd + "/testdata/vendor",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion: "3.4.0",
						IncludePaths:    []string{},
					},
					Lint: settings.LintConfig{
						IDs:                 []string{},
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
				DirPath:     cwd,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/a/d": []*ProtoFile{
						{
							Path:        cwd + "/testdata/a/d/file.proto",
							DisplayPath: "testdata/a/d/file.proto",
						},
						{
							Path:        cwd + "/testdata/a/d/file2.proto",
							DisplayPath: "testdata/a/d/file2.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/a/d",
					ExcludePrefixes: []string{
						cwd + "/testdata/a/d/file3.proto",
						cwd + "/testdata/a/d/vendor",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion: "3.2.0",
						IncludePaths:    []string{},
					},
					Lint: settings.LintConfig{
						IDs:                 []string{},
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
				DirPath:     cwd,
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/b": []*ProtoFile{
						{
							Path:        cwd + "/testdata/b/file.proto",
							DisplayPath: "testdata/b/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/b",
					ExcludePrefixes: []string{
						cwd + "/testdata/b/g/h",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion: "3.3.0",
						IncludePaths:    []string{},
					},
					Lint: settings.LintConfig{
						IDs:                 []string{},
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

func TestProtoSetProviderGetForDirCwdSubRel(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	protoSetProvider := newTestProtoSetProvider(t)
	protoSets, err := protoSetProvider.GetForDir(cwd, "testdata/d/g")
	require.NoError(t, err)
	require.Equal(
		t,
		[]*ProtoSet{
			&ProtoSet{
				WorkDirPath: cwd,
				DirPath:     cwd + "/testdata/d/g",
				DirPathToFiles: map[string][]*ProtoFile{
					cwd + "/testdata/d": []*ProtoFile{
						{
							Path:        cwd + "/testdata/d/file.proto",
							DisplayPath: "testdata/d/file.proto",
						},
					},
					cwd + "/testdata/d/g/h": []*ProtoFile{
						{
							Path:        cwd + "/testdata/d/g/h/file.proto",
							DisplayPath: "testdata/d/g/h/file.proto",
						},
					},
				},
				Config: settings.Config{
					DirPath: cwd + "/testdata/d",
					ExcludePrefixes: []string{
						cwd + "/testdata/d/vendor",
					},
					Compile: settings.CompileConfig{
						ProtobufVersion: "3.3.0",
						IncludePaths:    []string{},
					},
					Lint: settings.LintConfig{
						IDs:                 []string{},
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

func newTestProtoSetProvider(t *testing.T) ProtoSetProvider {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	return NewProtoSetProvider(ProtoSetProviderWithLogger(logger))
}
