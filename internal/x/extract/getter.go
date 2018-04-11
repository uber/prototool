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

package extract

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"go.uber.org/zap"
)

type getter struct {
	logger *zap.Logger
}

func newGetter(options ...GetterOption) *getter {
	getter := &getter{
		logger: zap.NewNop(),
	}
	for _, option := range options {
		option(getter)
	}
	return getter
}

func (g *getter) GetField(fileDescriptorSets []*descriptor.FileDescriptorSet, path string) (*Field, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("empty path")
	}
	if path[0] == '.' {
		path = path[1:]
	}
	split := strings.Split(path, ".")
	if len(split) < 2 {
		return nil, fmt.Errorf("no field for path %s", path)
	}
	message, err := g.GetMessage(fileDescriptorSets, strings.Join(split[0:len(split)-1], "."))
	if err != nil {
		return nil, err
	}
	var foundFieldDescriptorProto *descriptor.FieldDescriptorProto
	for _, fieldDescriptorProto := range append(message.GetField(), message.GetExtension()...) {
		if fieldDescriptorProto.GetName() == split[len(split)-1] {
			if foundFieldDescriptorProto != nil {
				return nil, fmt.Errorf("duplicate fields for path %s", path)
			}
			foundFieldDescriptorProto = fieldDescriptorProto
		}
	}
	if foundFieldDescriptorProto == nil {
		return nil, fmt.Errorf("no field for path %s", path)
	}
	return &Field{
		FieldDescriptorProto: foundFieldDescriptorProto,
		FullyQualifiedPath:   "." + path,
		DescriptorProto:      message.DescriptorProto,
		FileDescriptorProto:  message.FileDescriptorProto,
		FileDescriptorSet:    message.FileDescriptorSet,
	}, nil
}

func (g *getter) GetMessage(fileDescriptorSets []*descriptor.FileDescriptorSet, path string) (*Message, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("empty path")
	}
	if path[0] == '.' {
		path = path[1:]
	}
	var descriptorProto *descriptor.DescriptorProto
	var fileDescriptorProto *descriptor.FileDescriptorProto
	var fileDescriptorSet *descriptor.FileDescriptorSet
	for _, iFileDescriptorSet := range fileDescriptorSets {
		for _, iFileDescriptorProto := range iFileDescriptorSet.File {
			iDescriptorProto, err := findDescriptorProto(path, iFileDescriptorProto)
			if err != nil {
				return nil, err
			}
			if iDescriptorProto != nil {
				if descriptorProto != nil {
					return nil, fmt.Errorf("duplicate messages for path %s", path)
				}
				descriptorProto = iDescriptorProto
				fileDescriptorProto = iFileDescriptorProto
			}
		}
		// return first fileDescriptorSet that matches
		// as opposed to duplicate check within fileDescriptorSet, we easily could
		// have multiple fileDescriptorSets that match
		if descriptorProto != nil {
			fileDescriptorSet = iFileDescriptorSet
			break
		}
	}
	if descriptorProto == nil {
		return nil, fmt.Errorf("no message for path %s", path)
	}
	return &Message{
		DescriptorProto:     descriptorProto,
		FullyQualifiedPath:  "." + path,
		FileDescriptorProto: fileDescriptorProto,
		FileDescriptorSet:   fileDescriptorSet,
	}, nil
}

func (g *getter) GetService(fileDescriptorSets []*descriptor.FileDescriptorSet, path string) (*Service, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("empty path")
	}
	if path[0] == '.' {
		path = path[1:]
	}
	var serviceDescriptorProto *descriptor.ServiceDescriptorProto
	var fileDescriptorProto *descriptor.FileDescriptorProto
	var fileDescriptorSet *descriptor.FileDescriptorSet
	for _, iFileDescriptorSet := range fileDescriptorSets {
		for _, iFileDescriptorProto := range iFileDescriptorSet.File {
			iServiceDescriptorProto, err := findServiceDescriptorProto(path, iFileDescriptorProto)
			if err != nil {
				return nil, err
			}
			if iServiceDescriptorProto != nil {
				if serviceDescriptorProto != nil {
					return nil, fmt.Errorf("duplicate services for path %s", path)
				}
				serviceDescriptorProto = iServiceDescriptorProto
				fileDescriptorProto = iFileDescriptorProto
			}
		}
		// return first fileDescriptorSet that matches
		// as opposed to duplicate check within fileDescriptorSet, we easily could
		// have multiple fileDescriptorSets that match
		if serviceDescriptorProto != nil {
			fileDescriptorSet = iFileDescriptorSet
			break
		}
	}
	if serviceDescriptorProto == nil {
		return nil, fmt.Errorf("no service for path %s", path)
	}
	return &Service{
		ServiceDescriptorProto: serviceDescriptorProto,
		FullyQualifiedPath:     "." + path,
		FileDescriptorProto:    fileDescriptorProto,
		FileDescriptorSet:      fileDescriptorSet,
	}, nil
}

// TODO: we don't actually do full path resolution per the descriptor.proto spec
// https://github.com/google/protobuf/blob/master/src/google/protobuf/descriptor.proto#L185

func findDescriptorProto(path string, fileDescriptorProto *descriptor.FileDescriptorProto) (*descriptor.DescriptorProto, error) {
	if fileDescriptorProto.GetPackage() == "" {
		return nil, fmt.Errorf("no package on FileDescriptorProto")
	}
	if !strings.HasPrefix(path, fileDescriptorProto.GetPackage()) {
		return nil, nil
	}
	return findDescriptorProtoInSlice(path, fileDescriptorProto.GetPackage(), fileDescriptorProto.GetMessageType())
}

func findDescriptorProtoInSlice(path string, nestedName string, descriptorProtos []*descriptor.DescriptorProto) (*descriptor.DescriptorProto, error) {
	var foundDescriptorProto *descriptor.DescriptorProto
	for _, descriptorProto := range descriptorProtos {
		if descriptorProto.GetName() == "" {
			return nil, fmt.Errorf("no name on DescriptorProto")
		}
		fullName := nestedName + "." + descriptorProto.GetName()
		if path == fullName {
			if foundDescriptorProto != nil {
				return nil, fmt.Errorf("duplicate messages for path %s", path)
			}
			foundDescriptorProto = descriptorProto
		}
		nestedFoundDescriptorProto, err := findDescriptorProtoInSlice(path, fullName, descriptorProto.GetNestedType())
		if err != nil {
			return nil, err
		}
		if nestedFoundDescriptorProto != nil {
			if foundDescriptorProto != nil {
				return nil, fmt.Errorf("duplicate messages for path %s", path)
			}
			foundDescriptorProto = nestedFoundDescriptorProto
		}
	}
	return foundDescriptorProto, nil
}

func findServiceDescriptorProto(path string, fileDescriptorProto *descriptor.FileDescriptorProto) (*descriptor.ServiceDescriptorProto, error) {
	if fileDescriptorProto.GetPackage() == "" {
		return nil, fmt.Errorf("no package on FileDescriptorProto")
	}
	if !strings.HasPrefix(path, fileDescriptorProto.GetPackage()) {
		return nil, nil
	}
	var foundServiceDescriptorProto *descriptor.ServiceDescriptorProto
	for _, serviceDescriptorProto := range fileDescriptorProto.GetService() {
		if fileDescriptorProto.GetPackage()+"."+serviceDescriptorProto.GetName() == path {
			if foundServiceDescriptorProto != nil {
				return nil, fmt.Errorf("duplicate services for path %s", path)
			}
			foundServiceDescriptorProto = serviceDescriptorProto
		}
	}
	return foundServiceDescriptorProto, nil
}
