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

package format

import (
	"fmt"
	"strings"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

var _ proto.Visitor = &mainVisitor{}

type mainVisitor struct {
	*baseVisitor

	isProto2          bool
	haveHitNonComment bool
	parent            proto.Visitee
}

func newMainVisitor(isProto2 bool) *mainVisitor {
	return &mainVisitor{isProto2: isProto2, baseVisitor: newBaseVisitor()}
}

func (v *mainVisitor) Do() []*text.Failure {
	return v.Failures
}

func (v *mainVisitor) VisitMessage(element *proto.Message) {
	v.haveHitNonComment = true
	v.PComment(element.Comment)
	prefix := "message "
	if element.IsExtend {
		prefix = "extend "
	}
	if len(element.Elements) == 0 {
		v.P(prefix, element.Name, " {}")
		v.P()
		return
	}
	v.P(prefix, element.Name, " {")
	v.In()
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Elements {
		child.Accept(v)
	}
	v.parent = originalParent
	v.Out()
	v.P("}")
	if v.parent == nil {
		v.P()
	}
}

func (v *mainVisitor) VisitService(element *proto.Service) {
	v.haveHitNonComment = true
	v.PComment(element.Comment)
	if len(element.Elements) == 0 {
		v.P("service ", element.Name, " {}")
		v.P()
		return
	}
	v.P("service ", element.Name, " {")
	v.In()
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Elements {
		child.Accept(v)
	}
	v.parent = originalParent
	v.Out()
	v.P("}")
	v.P()
}

func (v *mainVisitor) VisitSyntax(element *proto.Syntax) {
	// done in first pass visitor
	v.haveHitNonComment = true
}

func (v *mainVisitor) VisitPackage(element *proto.Package) {
	// done in first pass visitor
	v.haveHitNonComment = true
}

func (v *mainVisitor) VisitOption(element *proto.Option) {
	v.haveHitNonComment = true
	// file options done in first pass visitor
	if v.parent == nil {
		return
	}
	switch v.parent.(type) {
	case (*proto.Enum):
		v.POptions(element)
	case (*proto.Message):
		v.POptions(element)
	case (*proto.Oneof):
		v.POptions(element)
	case (*proto.Service):
		v.POptions(element)
	default:
		v.AddFailure(element.Position, "unhandled child option")
	}
}

func (v *mainVisitor) VisitImport(element *proto.Import) {
	// done in first pass visitor
	v.haveHitNonComment = true
}

func (v *mainVisitor) VisitNormalField(element *proto.NormalField) {
	v.haveHitNonComment = true
	prefix := ""
	if v.isProto2 {
		// technically these are only set if the file is proto2
		// but doing this just to make sure
		if element.Required {
			prefix = "required "
		} else {
			prefix = "optional "
		}
	}
	if element.Repeated {
		prefix = "repeated "
	}
	v.PField(prefix, element.Type, element.Field)
}

func (v *mainVisitor) VisitEnumField(element *proto.EnumField) {
	v.haveHitNonComment = true
	v.PEnumField(element)
}

func (v *mainVisitor) VisitEnum(element *proto.Enum) {
	v.haveHitNonComment = true
	v.PComment(element.Comment)
	if len(element.Elements) == 0 {
		v.P("enum ", element.Name, " {}")
		v.P()
		return
	}
	v.P("enum ", element.Name, " {")
	v.In()
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Elements {
		child.Accept(v)
	}
	v.parent = originalParent
	v.Out()
	v.P("}")
	if v.parent == nil {
		v.P()
	}
}

func (v *mainVisitor) VisitComment(element *proto.Comment) {
	if v.haveHitNonComment {
		v.PComment(element)
		v.P()
	}
}

func (v *mainVisitor) VisitOneof(element *proto.Oneof) {
	v.haveHitNonComment = true
	v.PComment(element.Comment)
	if len(element.Elements) == 0 {
		// TODO: is this even legal?
		v.P("oneof ", element.Name, " {}")
		return
	}
	v.P("oneof ", element.Name, " {")
	v.In()
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Elements {
		child.Accept(v)
	}
	v.parent = originalParent
	v.Out()
	v.P("}")
}

func (v *mainVisitor) VisitOneofField(element *proto.OneOfField) {
	v.haveHitNonComment = true
	v.PField("", element.Type, element.Field)
}

func (v *mainVisitor) VisitReserved(element *proto.Reserved) {
	v.haveHitNonComment = true
	if len(element.Ranges) > 0 && len(element.FieldNames) > 0 {
		v.AddFailure(element.Position, "reserved had both integer ranges and field names which is unexpected")
		return
	}
	v.PComment(element.Comment)
	if len(element.Ranges) > 0 {
		rangeStrings := make([]string, len(element.Ranges))
		for i, r := range element.Ranges {
			rangeStrings[i] = r.SourceRepresentation()
		}
		v.PWithInlineComment(element.InlineComment, "reserved ", strings.Join(rangeStrings, ", "), ";")
		return
	}
	if len(element.FieldNames) > 0 {
		fieldNameStrings := make([]string, len(element.FieldNames))
		for i, fieldName := range element.FieldNames {
			fieldNameStrings[i] = `"` + fieldName + `"`
		}
		v.PWithInlineComment(element.InlineComment, "reserved ", strings.Join(fieldNameStrings, ", "), ";")
	}
}

func (v *mainVisitor) VisitRPC(element *proto.RPC) {
	v.haveHitNonComment = true
	v.PComment(element.Comment)
	requestStream := ""
	if element.StreamsRequest {
		requestStream = "stream "
	}
	responseStream := ""
	if element.StreamsReturns {
		responseStream = "stream "
	}
	if len(element.Options) == 0 {
		v.PWithInlineComment(element.InlineComment, "rpc ", element.Name, "(", requestStream, element.RequestType, ") returns (", responseStream, element.ReturnsType, ");")
		return
	}
	v.P("rpc ", element.Name, "(", requestStream, element.RequestType, ") returns (", responseStream, element.ReturnsType, ") {")
	v.In()
	v.POptions(element.Options...)
	v.Out()
	v.PWithInlineComment(element.InlineComment, "}")
}

func (v *mainVisitor) VisitMapField(element *proto.MapField) {
	v.haveHitNonComment = true
	v.PField("", fmt.Sprintf("map<%s, %s>", element.KeyType, element.Type), element.Field)
}

func (v *mainVisitor) VisitGroup(element *proto.Group) {
	v.haveHitNonComment = true
	v.PComment(element.Comment)
	prefix := ""
	// TODO: required and repeated not handled yet, add when handled
	if element.Optional {
		prefix = "optional "
	}
	if len(element.Elements) == 0 {
		v.P(prefix, "group ", element.Name, " = ", element.Sequence, " {}")
		return
	}
	v.P(prefix, "group ", element.Name, " = ", element.Sequence, " {")
	v.In()
	originalParent := v.parent
	v.parent = element
	for _, child := range element.Elements {
		child.Accept(v)
	}
	v.parent = originalParent
	v.Out()
	v.P("}")
}

func (v *mainVisitor) VisitExtensions(element *proto.Extensions) {
	v.haveHitNonComment = true
	v.PComment(element.Comment)
	rangeStrings := make([]string, len(element.Ranges))
	for i, r := range element.Ranges {
		rangeStrings[i] = r.SourceRepresentation()
	}
	v.PWithInlineComment(element.InlineComment, "extensions ", strings.Join(rangeStrings, ", "), ";")
}
