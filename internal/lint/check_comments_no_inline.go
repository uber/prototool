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

var commentsNoInlineLinter = NewLinter(
	"COMMENTS_NO_INLINE",
	"Verifies that there are no inline comments.",
	checkCommentsNoInline,
)

func checkCommentsNoInline(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&commentsNoInlineVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type commentsNoInlineVisitor struct {
	baseAddVisitor
}

func (v commentsNoInlineVisitor) VisitMessage(element *proto.Message) {
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoInlineVisitor) VisitService(element *proto.Service) {
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoInlineVisitor) VisitSyntax(element *proto.Syntax) {
	v.checkInlineComment("syntax", element.InlineComment)
}

func (v commentsNoInlineVisitor) VisitPackage(element *proto.Package) {
	v.checkInlineComment("packages", element.InlineComment)
}

func (v commentsNoInlineVisitor) VisitOption(element *proto.Option) {
	v.checkInlineComment("options", element.InlineComment)
}

func (v commentsNoInlineVisitor) VisitImport(element *proto.Import) {
	v.checkInlineComment("imports", element.InlineComment)
}

func (v commentsNoInlineVisitor) VisitNormalField(element *proto.NormalField) {
	v.checkInlineComment("fields", element.InlineComment)
	for _, child := range element.Options {
		child.Accept(v)
	}
}

func (v commentsNoInlineVisitor) VisitEnumField(element *proto.EnumField) {
	v.checkInlineComment("enum values", element.InlineComment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoInlineVisitor) VisitEnum(element *proto.Enum) {
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoInlineVisitor) VisitOneof(element *proto.Oneof) {
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoInlineVisitor) VisitOneofField(element *proto.OneOfField) {
	v.checkInlineComment("oneofs", element.InlineComment)
	for _, child := range element.Options {
		child.Accept(v)
	}
}

func (v commentsNoInlineVisitor) VisitReserved(element *proto.Reserved) {
	v.checkInlineComment("reserved", element.InlineComment)
}

func (v commentsNoInlineVisitor) VisitRPC(element *proto.RPC) {
	v.checkInlineComment("RPCs", element.InlineComment)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoInlineVisitor) VisitMapField(element *proto.MapField) {
	v.checkInlineComment("fields", element.InlineComment)
	for _, child := range element.Options {
		child.Accept(v)
	}
}

func (v commentsNoInlineVisitor) VisitGroup(element *proto.Group) {
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v commentsNoInlineVisitor) VisitExtensions(element *proto.Extensions) {
	v.checkInlineComment("extensions", element.InlineComment)
}

func (v commentsNoInlineVisitor) checkInlineComment(t string, inlineComment *proto.Comment) {
	if inlineComment != nil {
		v.AddFailuref(inlineComment.Position, "Inline comments are not allowed on %s, only comment above the type.", t)
	}
}
