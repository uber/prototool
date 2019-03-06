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
	"github.com/uber/prototool/internal/text"
)

var enumsHaveCommentsLinter = NewLinter(
	"ENUMS_HAVE_COMMENTS",
	`Verifies that all enums have a comment of the form "// EnumName ...".`,
	checkEnumsHaveComments,
)

func checkEnumsHaveComments(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(enumsHaveCommentsVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type enumsHaveCommentsVisitor struct {
	baseAddVisitor
}

func (v enumsHaveCommentsVisitor) VisitMessage(message *proto.Message) {
	// for nested enums
	for _, child := range message.Elements {
		child.Accept(v)
	}
}

func (v enumsHaveCommentsVisitor) VisitEnum(enum *proto.Enum) {
	if !hasGolangStyleComment(enum.Comment, enum.Name) {
		v.AddFailuref(enum.Position, `Enum %q needs a comment of the form "// %s ..."`, enum.Name, enum.Name)
	}
}
