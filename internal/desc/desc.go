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

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// SortFileDescriptorSet sorts a FileDescriptorSet for github.com/jhump/protoreflect
// by returning a new FileDescriptorSet with the given FileDescriptorProto at the end.
// This also verifies that all FileDescriptorProto names are unique and the name of the
// FileDescriptorProto is within the FileDescriptorSet.
func SortFileDescriptorSet(fileDescriptorSet *descriptor.FileDescriptorSet, fileDescriptorProto *descriptor.FileDescriptorProto) (*descriptor.FileDescriptorSet, error) {
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
