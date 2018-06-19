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

package reflect

import (
	"fmt"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/uber/prototool/internal/extract"
	intdesc "github.com/uber/prototool/internal/x/desc"
	"go.uber.org/zap"
)

type handler struct {
	logger *zap.Logger

	getter extract.Getter
}

func newHandler(options ...HandlerOption) *handler {
	handler := &handler{
		logger: zap.NewNop(),
	}
	for _, option := range options {
		option(handler)
	}
	// TODO(pedge): composition
	handler.getter = extract.NewGetter(
		extract.GetterWithLogger(handler.logger),
	)
	return handler
}

func (h *handler) BinaryToJSON(fileDescriptorSets []*descriptor.FileDescriptorSet, messagePath string, binaryData []byte) ([]byte, error) {
	dynamicMessage, err := h.getDynamicMessage(fileDescriptorSets, messagePath)
	if err != nil {
		return nil, err
	}
	if err := dynamicMessage.Unmarshal(binaryData); err != nil {
		return nil, err
	}
	return dynamicMessage.MarshalJSON()
}

func (h *handler) JSONToBinary(fileDescriptorSets []*descriptor.FileDescriptorSet, messagePath string, jsonData []byte) ([]byte, error) {
	dynamicMessage, err := h.getDynamicMessage(fileDescriptorSets, messagePath)
	if err != nil {
		return nil, err
	}
	if err := dynamicMessage.UnmarshalJSON(jsonData); err != nil {
		return nil, err
	}
	return dynamicMessage.Marshal()
}

func (h *handler) getDynamicMessage(fileDescriptorSets []*descriptor.FileDescriptorSet, messagePath string) (*dynamic.Message, error) {
	message, err := h.getter.GetMessage(fileDescriptorSets, messagePath)
	if err != nil {
		return nil, err
	}
	fileDescriptorSet, err := intdesc.SortFileDescriptorSet(message.FileDescriptorSet, message.FileDescriptorProto)
	if err != nil {
		return nil, err
	}
	fileDescriptor, err := desc.CreateFileDescriptorFromSet(fileDescriptorSet)
	if err != nil {
		return nil, err
	}
	if len(message.FullyQualifiedPath) == 0 || message.FullyQualifiedPath[0] != '.' {
		return nil, fmt.Errorf("malformed FullyQualifiedPath: %s", message.FullyQualifiedPath)
	}
	messageDescriptor := fileDescriptor.FindMessage(message.FullyQualifiedPath[1:])
	if messageDescriptor == nil {
		return nil, fmt.Errorf("no MessageDescriptor for path %s", message.FullyQualifiedPath)
	}
	return dynamic.NewMessage(messageDescriptor), nil
}
