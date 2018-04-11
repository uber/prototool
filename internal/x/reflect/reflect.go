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
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"go.uber.org/zap"
)

// Handler handles reflection.
type Handler interface {
	BinaryToJSON(fileDescriptorSets []*descriptor.FileDescriptorSet, messagePath string, binaryData []byte) ([]byte, error)
	JSONToBinary(fileDescriptorSets []*descriptor.FileDescriptorSet, messagePath string, jsonData []byte) ([]byte, error)
}

// HandlerOption is an option for a new Handler.
type HandlerOption func(*handler)

// HandlerWithLogger returns a HandlerOption that uses the given logger.
//
// The default is to use zap.NewNop().
func HandlerWithLogger(logger *zap.Logger) HandlerOption {
	return func(handler *handler) {
		handler.logger = logger
	}
}

// NewHandler returns a new Handler.
func NewHandler(options ...HandlerOption) Handler {
	return newHandler(options...)
}
