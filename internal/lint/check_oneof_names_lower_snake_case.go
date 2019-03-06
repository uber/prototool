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

package lint

import (
	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/strs"
	"github.com/uber/prototool/internal/text"
)

var oneofNamesLowerSnakeCaseLinter = NewLinter(
	"ONEOF_NAMES_LOWER_SNAKE_CASE",
	"Verifies that all oneof names are lower_snake_case.",
	checkOneofNamesLowerSnakeCase,
)

func checkOneofNamesLowerSnakeCase(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(oneofNamesLowerSnakeCaseVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type oneofNamesLowerSnakeCaseVisitor struct {
	baseAddVisitor
}

func (v oneofNamesLowerSnakeCaseVisitor) VisitMessage(message *proto.Message) {
	for _, element := range message.Elements {
		element.Accept(v)
	}
}

func (v oneofNamesLowerSnakeCaseVisitor) VisitOneof(oneof *proto.Oneof) {
	if !strs.IsLowerSnakeCase(oneof.Name) {
		v.AddFailuref(oneof.Position, "Oneof name %q must be lower_snake_case.", oneof.Name)
	}
}
