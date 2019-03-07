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

package desc

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/golang/protobuf/proto"
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

// MergeFileDescriptorSets merges the given FileDescriptorSets, checking that files
// with the same name have the same content.
func MergeFileDescriptorSets(fileDescriptorSets []*descriptor.FileDescriptorSet) (*descriptor.FileDescriptorSet, error) {
	result := &descriptor.FileDescriptorSet{
		File: make([]*descriptor.FileDescriptorProto, 0),
	}
	if len(fileDescriptorSets) == 0 {
		return result, nil
	}
	nameToSum := make(map[string][]byte)
	for _, fileDescriptorSet := range fileDescriptorSets {
		if fileDescriptorSet == nil {
			return nil, fmt.Errorf("nil FileDescriptorSet")
		}
		for _, fileDescriptorProto := range fileDescriptorSet.File {
			if fileDescriptorProto == nil {
				return nil, fmt.Errorf("nil FileDescriptorProto")
			}
			name := fileDescriptorProto.GetName()
			if name == "" {
				return nil, fmt.Errorf("no name on FileDescriptorProto")
			}
			sum, err := sha256ProtoMessage(fileDescriptorProto)
			if err != nil {
				return nil, err
			}
			if existingSum, ok := nameToSum[name]; ok {
				if !bytes.Equal(existingSum, sum) {
					return nil, fmt.Errorf("mismatched sums %x %x for file %q", existingSum, sum, name)
				}
			} else {
				nameToSum[name] = sum
				result.File = append(result.File, fileDescriptorProto)
			}
		}
	}
	sort.Slice(result.File, func(i int, j int) bool {
		return result.File[i].GetName() < result.File[j].GetName()
	})
	return result, nil
}

func sha256ProtoMessage(message proto.Message) ([]byte, error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	array := sha256.Sum256(data)
	return array[:], nil
}
