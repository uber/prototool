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
	"text/scanner"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

var commentsNoCStyleLinter = NewLinter(
	"COMMENTS_NO_C_STYLE",
	"Verifies that there are no /* c-style */ comments.",
	checkCommentsNoCStyle,
)

func checkCommentsNoCStyle(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&commentsNoCStyleVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type commentsNoCStyleVisitor struct {
	baseAddVisitor
}

func (v commentsNoCStyleVisitor) VisitMessage(element *proto.Message) {
	v.checkComments(element.Position, element.Comment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoCStyleVisitor) VisitService(element *proto.Service) {
	v.checkComments(element.Position, element.Comment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoCStyleVisitor) VisitSyntax(element *proto.Syntax) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
}

func (v commentsNoCStyleVisitor) VisitPackage(element *proto.Package) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
}

func (v commentsNoCStyleVisitor) VisitOption(element *proto.Option) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
}

func (v commentsNoCStyleVisitor) VisitImport(element *proto.Import) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
}

func (v commentsNoCStyleVisitor) VisitNormalField(element *proto.NormalField) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
	for _, child := range element.Options {
		child.Accept(v)
	}
}

func (v commentsNoCStyleVisitor) VisitEnumField(element *proto.EnumField) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoCStyleVisitor) VisitEnum(element *proto.Enum) {
	v.checkComments(element.Position, element.Comment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoCStyleVisitor) VisitComment(element *proto.Comment) {
	v.checkComments(element.Position, element)
}

func (v commentsNoCStyleVisitor) VisitOneof(element *proto.Oneof) {
	v.checkComments(element.Position, element.Comment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoCStyleVisitor) VisitOneofField(element *proto.OneOfField) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
	for _, child := range element.Options {
		child.Accept(v)
	}
}

func (v commentsNoCStyleVisitor) VisitReserved(element *proto.Reserved) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
}

func (v commentsNoCStyleVisitor) VisitRPC(element *proto.RPC) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoCStyleVisitor) VisitMapField(element *proto.MapField) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
	for _, child := range element.Options {
		child.Accept(v)
	}
}

func (v commentsNoCStyleVisitor) VisitGroup(element *proto.Group) {
	v.checkComments(element.Position, element.Comment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoCStyleVisitor) VisitExtensions(element *proto.Extensions) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
}

func (v commentsNoCStyleVisitor) checkComments(position scanner.Position, comments ...*proto.Comment) {
	for _, comment := range comments {
		if comment != nil {
			if comment.Cstyle {
				v.AddFailuref(position, "C-Style comments are not allowed.")
			}
		}
	}
}
