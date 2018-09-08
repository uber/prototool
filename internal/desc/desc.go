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

package desc

import (
	"fmt"
	"sort"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// SortFileDescriptorSetAtEnd sorts a FileDescriptorSet for github.com/jhump/protoreflect
// by returning a new FileDescriptorSet with the given FileDescriptorProto at the end.
// This also verifies that all FileDescriptorProto names are unique and the name of the
// FileDescriptorProto is within the FileDescriptorSet.
func SortFileDescriptorSetAtEnd(fileDescriptorSet *descriptor.FileDescriptorSet, fileDescriptorProto *descriptor.FileDescriptorProto) (*descriptor.FileDescriptorSet, error) {
	// best-effort checks
	names := make(map[string]struct{}, len(fileDescriptorSet.File))
	for _, iFileDescriptorProto := range fileDescriptorSet.File {
		if iFileDescriptorProto.GetName() == "" {
			return nil, fmt.Errorf("no name on FileDescriptorProto")
		}
		if _, ok := names[iFileDescriptorProto.GetName()]; ok {
			return nil, fmt.Errorf("duplicate FileDescriptorProto in FileDescriptorSet: %s", iFileDescriptorProto.GetName())
		}
		names[iFileDescriptorProto.GetName()] = struct{}{}
	}
	if _, ok := names[fileDescriptorProto.GetName()]; !ok {
		return nil, fmt.Errorf("no FileDescriptorProto named %s in FileDescriptorSet with names %v", fileDescriptorProto.GetName(), names)
	}
	newFileDescriptorSet := &descriptor.FileDescriptorSet{}
	for _, iFileDescriptorProto := range fileDescriptorSet.File {
		if iFileDescriptorProto.GetName() != fileDescriptorProto.GetName() {
			newFileDescriptorSet.File = append(newFileDescriptorSet.File, iFileDescriptorProto)
		}
	}
	newFileDescriptorSet.File = append(newFileDescriptorSet.File, fileDescriptorProto)
	return newFileDescriptorSet, nil
}

// SortFileDescriptorSet sorts a FileDescriptorSet in place by using its
// relative import path and file name retrieved from GetName().
//
// This assumes that the give FileDescriptorSet is valid.
func SortFileDescriptorSet(fileDescriptorSet *descriptor.FileDescriptorSet) {
	sort.Slice(fileDescriptorSet.File, func(i, j int) bool {
		return fileDescriptorSet.File[i].GetName() < fileDescriptorSet.File[j].GetName()
	})
}

// MergeFileDescriptorSets deduplicates and sorts multiple file descriptor sets
// into a single set.
func MergeFileDescriptorSets(sets []*descriptor.FileDescriptorSet) *descriptor.FileDescriptorSet {
	// Here we rely soley on GetName() to uniquely identify each
	// FileDescriptorProto. GetName() returns the unique proto file name,
	// including the import path from the current directory.
	//
	// eg: GetName() would return "bar/bar.proto" and "foo/foo.proto",
	// respectively, for bar.proto and foo.proto in the following directory
	// structure.
	//
	//  .
	//  ├── bar
	//  │   └── bar.proto
	//  ├── foo
	//  │   └── foo.proto
	//  └── prototool.yaml
	files := map[string]*descriptor.FileDescriptorProto{}
	for _, fileDescriptorSet := range sets {
		for _, fileDescriptorProto := range fileDescriptorSet.File {
			name := fileDescriptorProto.GetName()
			if _, ok := files[name]; !ok {
				files[name] = fileDescriptorProto
			}
		}
	}

	unifiedFiles := make([]*descriptor.FileDescriptorProto, 0, len(files))
	for _, f := range files {
		unifiedFiles = append(unifiedFiles, f)
	}

	set := &descriptor.FileDescriptorSet{File: unifiedFiles}
	SortFileDescriptorSet(set)
	return set
}
