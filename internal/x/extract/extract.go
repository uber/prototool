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
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"go.uber.org/zap"
)

// Field is an extracted field.
type Field struct {
	*descriptor.FieldDescriptorProto

	FullyQualifiedPath  string
	DescriptorProto     *descriptor.DescriptorProto
	FileDescriptorProto *descriptor.FileDescriptorProto
	FileDescriptorSet   *descriptor.FileDescriptorSet
}

// Message is an extracted message.
type Message struct {
	*descriptor.DescriptorProto

	FullyQualifiedPath  string
	FileDescriptorProto *descriptor.FileDescriptorProto
	FileDescriptorSet   *descriptor.FileDescriptorSet
}

// Service is an extracted service.
type Service struct {
	*descriptor.ServiceDescriptorProto

	FullyQualifiedPath  string
	FileDescriptorProto *descriptor.FileDescriptorProto
	FileDescriptorSet   *descriptor.FileDescriptorSet
}

// Getter extracts elements.
//
// Paths can begin with ".".
// The first FileDescriptorSet with a match will be returned.
type Getter interface {
	// Get the field that matches the path.
	GetField(fileDescriptorSets []*descriptor.FileDescriptorSet, path string) (*Field, error)
	// Get the message that matches the path.
	GetMessage(fileDescriptorSets []*descriptor.FileDescriptorSet, path string) (*Message, error)
	// Get the service that matches the path.
	GetService(fileDescriptorSets []*descriptor.FileDescriptorSet, path string) (*Service, error)
}

// GetterOption is an option for a new Getter.
type GetterOption func(*getter)

// GetterWithLogger returns a GetterOption that uses the given logger.
//
// The default is to use zap.NewNop().
func GetterWithLogger(logger *zap.Logger) GetterOption {
	return func(getter *getter) {
		getter.logger = logger
	}
}

// NewGetter returns a new Getter.
func NewGetter(options ...GetterOption) Getter {
	return newGetter(options...)
}
