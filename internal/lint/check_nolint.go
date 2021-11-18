// Copyright (c) 2020 Uber Technologies, Inc.
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
	"strings"
	"text/scanner"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

var nolintLinter = NewLinter(
	"NOLINT",
	"Checks lines do not need to lint.",
	checkNolint,
)

func checkNolint(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&nolintVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type nolintVisitor struct {
	baseAddVisitor
}

func (v nolintVisitor) VisitMessage(element *proto.Message) {
	v.checkComments(element.Position, element.Comment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v nolintVisitor) VisitService(element *proto.Service) {
	v.checkComments(element.Position, element.Comment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v nolintVisitor) VisitSyntax(element *proto.Syntax) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
}

func (v nolintVisitor) VisitPackage(element *proto.Package) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
}

func (v nolintVisitor) VisitOption(element *proto.Option) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
}

func (v nolintVisitor) VisitImport(element *proto.Import) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
}

func (v nolintVisitor) VisitNormalField(element *proto.NormalField) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
	for _, child := range element.Options {
		child.Accept(v)
	}
}

func (v nolintVisitor) VisitEnumField(element *proto.EnumField) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v nolintVisitor) VisitEnum(element *proto.Enum) {
	v.checkComments(element.Position, element.Comment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v nolintVisitor) VisitComment(element *proto.Comment) {
	v.checkComments(element.Position, element)
}

func (v nolintVisitor) VisitOneof(element *proto.Oneof) {
	v.checkComments(element.Position, element.Comment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v nolintVisitor) VisitOneofField(element *proto.OneOfField) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
	for _, child := range element.Options {
		child.Accept(v)
	}
}

func (v nolintVisitor) VisitReserved(element *proto.Reserved) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
}

func (v nolintVisitor) VisitRPC(element *proto.RPC) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v nolintVisitor) VisitMapField(element *proto.MapField) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
	for _, child := range element.Options {
		child.Accept(v)
	}
}

func (v nolintVisitor) VisitGroup(element *proto.Group) {
	v.checkComments(element.Position, element.Comment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v nolintVisitor) VisitExtensions(element *proto.Extensions) {
	v.checkComments(element.Position, element.Comment, element.InlineComment)
}

func (v nolintVisitor) checkComments(position scanner.Position, comments ...*proto.Comment) {
	for _, comment := range comments {
		if comment != nil {
			for _, line := range comment.Lines {
				if strings.Contains(line, "nolint") {
					v.add(text.NewFailuref(position, "", ""))
				}
			}
		}
	}
}
