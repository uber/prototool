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
	"github.com/uber/prototool/internal/file"
	"github.com/uber/prototool/internal/text"
)

var messageFieldsNotFloatsLinter = NewSuppressableLinter(
	"MESSAGE_FIELDS_NOT_FLOATS",
	`Verifies that all message fields are not floats.`,
	"floats",
	checkMessageFieldsNotFloats,
)

func checkMessageFieldsNotFloats(add func(*file.ProtoSet, *proto.Comment, *text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&messageFieldsNotFloatsVisitor{baseAddSuppressableVisitor: newBaseAddSuppressableVisitor(add)}, descriptors)
}

type messageFieldsNotFloatsVisitor struct {
	*baseAddSuppressableVisitor
}

func (v *messageFieldsNotFloatsVisitor) VisitMessage(message *proto.Message) {
	for _, element := range message.Elements {
		element.Accept(v)
	}
}

func (v *messageFieldsNotFloatsVisitor) VisitOneof(oneof *proto.Oneof) {
	for _, element := range oneof.Elements {
		element.Accept(v)
	}
}

func (v *messageFieldsNotFloatsVisitor) VisitNormalField(field *proto.NormalField) {
	v.checkNotFloat(field.Field)
}

func (v *messageFieldsNotFloatsVisitor) VisitOneofField(field *proto.OneOfField) {
	v.checkNotFloat(field.Field)
}

func (v *messageFieldsNotFloatsVisitor) VisitMapField(field *proto.MapField) {
	v.checkNotFloat(field.Field)
}

func (v *messageFieldsNotFloatsVisitor) checkNotFloat(field *proto.Field) {
	switch field.Type {
	case "float":
		v.AddFailuref(field.Comment, field.Position, `Field %q is a float and floats should generally not be used, consider using a double, or an int64 while representing your value in micros or nanos.`, field.Name)
	}
}
