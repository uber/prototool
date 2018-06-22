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

package format

import (
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/text"
	"go.uber.org/zap"
)

// Transformer transforms an input file into an output file.
type Transformer interface {
	// Transform transforms the data.
	//
	// Failures should never happen in the CLI tool as we run the files
	// through protoc first, but this is done because we want to verify
	// code correctness here and protect against the bad case.
	Transform(config settings.Config, filename string, data []byte) ([]byte, []*text.Failure, error)
}

// TransformerOption is an option for a new Transformer.
type TransformerOption func(*transformer)

// TransformerWithLogger returns a TransformerOption that uses the given logger.
//
// The default is to use zap.NewNop().
func TransformerWithLogger(logger *zap.Logger) TransformerOption {
	return func(transformer *transformer) {
		transformer.logger = logger
	}
}

// TransformerWithRewrite returns a TransformerOption that will update the file options
// go_package, java_package to match the package per the guidelines of the style guide.
func TransformerWithRewrite() TransformerOption {
	return func(transformer *transformer) {
		transformer.rewrite = true
	}
}

// NewTransformer returns a new Transformer.
func NewTransformer(options ...TransformerOption) Transformer {
	return newTransformer(options...)
}
