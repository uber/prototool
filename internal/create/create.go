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

package create

import "go.uber.org/zap"

// DefaultPackage is the default package to use in lieu of one being able to be
// derived.
const DefaultPackage = "uber.prototool.generated"

// Handler handles creation of Protobuf files from a template.
type Handler interface {
	// Create the files at the given filePaths.
	Create(filePaths ...string) error
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

// HandlerWithPackage returns a HandlerOption that uses the given package for
// new Protobuf files.
//
// The default is to derive this from the file path, or use DefaultPackage.
func HandlerWithPackage(pkg string) HandlerOption {
	return func(handler *handler) {
		handler.pkg = pkg
	}
}

// NewHandler returns a new Handler.
func NewHandler(options ...HandlerOption) Handler {
	return newHandler(options...)
}
