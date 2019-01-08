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
	"fmt"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

const messageFieldsNotFloatsSuppressableAnnotation = "floats"

var messageFieldsNotFloatsLinter = NewLinter(
	"MESSAGE_FIELDS_NOT_FLOATS",
	fmt.Sprintf(`Suppressable with "@suppresswarnings %s". Verifies that all message fields are not floats or doubles.`, messageFieldsNotFloatsSuppressableAnnotation),
	checkMessageFieldsNotFloats,
)

func checkMessageFieldsNotFloats(add func(*text.Failure), dirPath string, descriptors []*proto.Proto) error {
	return runVisitor(messageFieldsNotFloatsVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type messageFieldsNotFloatsVisitor struct {
	baseAddVisitor
}

func (v messageFieldsNotFloatsVisitor) VisitMessage(message *proto.Message) {
	for _, element := range message.Elements {
		element.Accept(v)
	}
}

func (v messageFieldsNotFloatsVisitor) VisitOneof(oneof *proto.Oneof) {
	for _, element := range oneof.Elements {
		element.Accept(v)
	}
}

func (v messageFieldsNotFloatsVisitor) VisitNormalField(field *proto.NormalField) {
	v.checkNotFloat(field.Field)
}

func (v messageFieldsNotFloatsVisitor) VisitOneofField(field *proto.OneOfField) {
	v.checkNotFloat(field.Field)
}

func (v messageFieldsNotFloatsVisitor) VisitMapField(field *proto.MapField) {
	v.checkNotFloat(field.Field)
}

func (v messageFieldsNotFloatsVisitor) checkNotFloat(field *proto.Field) {
	switch field.Type {
	case "double", "float":
		if isSuppressed(field.Comment, messageFieldsNotFloatsSuppressableAnnotation) {
			return
		}
		v.AddFailuref(field.Position, `Field %q is a float and floating point types should generally not be used, consider using an int64 while representing your value in micros or nanos. This can be suppressed by adding "@suppresswarnings %s" to the field comment.`, field.Name, messageFieldsNotFloatsSuppressableAnnotation)
	}
}
