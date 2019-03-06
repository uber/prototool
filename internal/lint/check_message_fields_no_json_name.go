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

var messageFieldsNoJSONNameLinter = NewLinter(
	"MESSAGE_FIELDS_NO_JSON_NAME",
	`Verifies that no message field has the "json_name" option set.`,
	checkMessageFieldsNoJSONName,
)

func checkMessageFieldsNoJSONName(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(messageFieldsNoJSONNameVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type messageFieldsNoJSONNameVisitor struct {
	baseAddVisitor
}

func (v messageFieldsNoJSONNameVisitor) VisitMessage(message *proto.Message) {
	for _, element := range message.Elements {
		element.Accept(v)
	}
}

func (v messageFieldsNoJSONNameVisitor) VisitOneof(oneof *proto.Oneof) {
	for _, element := range oneof.Elements {
		element.Accept(v)
	}
}

func (v messageFieldsNoJSONNameVisitor) VisitNormalField(field *proto.NormalField) {
	v.handleField(field.Field)
}

func (v messageFieldsNoJSONNameVisitor) VisitOneofField(field *proto.OneOfField) {
	v.handleField(field.Field)
}

func (v messageFieldsNoJSONNameVisitor) VisitMapField(field *proto.MapField) {
	v.handleField(field.Field)
}

func (v messageFieldsNoJSONNameVisitor) handleField(field *proto.Field) {
	for _, option := range field.Options {
		if option.Name == "json_name" {
			v.AddFailuref(field.Position, `Field %q cannot have the "json_name" option set.`, field.Name)
		}
	}
}
