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
	"fmt"

	"github.com/emicklei/proto"
	"go.uber.org/zap"
)

type logVisitor struct {
	*baseVisitor

	Logger *zap.Logger

	parent proto.Visitee
}

func newLogVisitor(logger *zap.Logger) *logVisitor {
	return &logVisitor{baseVisitor: newBaseVisitor(""), Logger: logger}
}

func (v *logVisitor) VisitMessage(element *proto.Message) {
	v.logVisitee(element)
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Elements {
		child.Accept(v)
	}
	v.parent = originalParent
}

func (v *logVisitor) VisitService(element *proto.Service) {
	v.logVisitee(element)
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Elements {
		child.Accept(v)
	}
	v.parent = originalParent
}

func (v *logVisitor) VisitSyntax(element *proto.Syntax) {
	v.logVisitee(element)
}

func (v *logVisitor) VisitPackage(element *proto.Package) {
	v.logVisitee(element)
}

func (v *logVisitor) VisitOption(element *proto.Option) {
	v.logVisitee(element)
}

func (v *logVisitor) VisitImport(element *proto.Import) {
	v.logVisitee(element)
}

func (v *logVisitor) VisitNormalField(element *proto.NormalField) {
	v.logVisitee(element, fmt.Sprintf("%+v", element.Field))
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Options {
		child.Accept(v)
	}
	v.parent = originalParent
}

func (v *logVisitor) VisitEnumField(element *proto.EnumField) {
	v.logVisitee(element)
	if element.ValueOption != nil {
		originalParent := v.parent
		v.parent = element
		element.ValueOption.Accept(v)
		v.parent = originalParent
	}
}

func (v *logVisitor) VisitEnum(element *proto.Enum) {
	v.logVisitee(element)
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Elements {
		child.Accept(v)
	}
	v.parent = originalParent
}

func (v *logVisitor) VisitComment(element *proto.Comment) {
	v.logVisitee(element)
}

func (v *logVisitor) VisitOneof(element *proto.Oneof) {
	v.logVisitee(element)
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Elements {
		child.Accept(v)
	}
	v.parent = originalParent
}

func (v *logVisitor) VisitOneofField(element *proto.OneOfField) {
	v.logVisitee(element)
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Options {
		child.Accept(v)
	}
	v.parent = originalParent
}

func (v *logVisitor) VisitReserved(element *proto.Reserved) {
	v.logVisitee(element)
}

func (v *logVisitor) VisitRPC(element *proto.RPC) {
	v.logVisitee(element)
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Options {
		child.Accept(v)
	}
	v.parent = originalParent
}

func (v *logVisitor) VisitMapField(element *proto.MapField) {
	v.logVisitee(element, fmt.Sprintf("%+v", element.Field))
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Options {
		child.Accept(v)
	}
	v.parent = originalParent
}

func (v *logVisitor) VisitGroup(element *proto.Group) {
	v.logVisitee(element)
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Elements {
		child.Accept(v)
	}
	v.parent = originalParent
}

func (v *logVisitor) VisitExtensions(element *proto.Extensions) {
	v.logVisitee(element)
}

func (v *logVisitor) logVisitee(visitee proto.Visitee, args ...interface{}) {
	v.Logger.Debug("",
		zap.String("type", fmt.Sprintf("%T", visitee)),
		zap.String("parent_type", fmt.Sprintf("%T", v.parent)),
		zap.String("element", fmt.Sprintf("%+v", visitee)),
		zap.String("args", fmt.Sprint(args...)))
}
