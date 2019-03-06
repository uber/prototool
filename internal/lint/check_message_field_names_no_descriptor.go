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
	"strings"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

var messageFieldNamesNoDescriptorLinter = NewLinter(
	"MESSAGE_FIELD_NAMES_NO_DESCRIPTOR",
	`Verifies that all message field names are not "descriptor", which results in a collision in Java-generated code.`,
	checkMessageFieldNamesNoDescriptor,
)

func checkMessageFieldNamesNoDescriptor(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(messageFieldNamesNoDescriptorVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type messageFieldNamesNoDescriptorVisitor struct {
	baseAddVisitor
}

func (v messageFieldNamesNoDescriptorVisitor) VisitMessage(message *proto.Message) {
	for _, element := range message.Elements {
		element.Accept(v)
	}
}

func (v messageFieldNamesNoDescriptorVisitor) VisitOneof(oneof *proto.Oneof) {
	for _, element := range oneof.Elements {
		element.Accept(v)
	}
}

func (v messageFieldNamesNoDescriptorVisitor) VisitNormalField(field *proto.NormalField) {
	v.handleField(field.Field)
}

func (v messageFieldNamesNoDescriptorVisitor) VisitOneofField(field *proto.OneOfField) {
	v.handleField(field.Field)
}

func (v messageFieldNamesNoDescriptorVisitor) VisitMapField(field *proto.MapField) {
	v.handleField(field.Field)
}

func (v messageFieldNamesNoDescriptorVisitor) handleField(field *proto.Field) {
	// Descriptor, _descriptor, descriptor_, __descriptor, descriptor__ are also invalid
	// https://github.com/protocolbuffers/protobuf/issues/5742
	if strings.Trim(strings.ToLower(field.Name), "_") == "descriptor" {
		v.AddFailuref(field.Position, `Field name %q causes a collision in Java-generated code.`, field.Name)
	}
}
