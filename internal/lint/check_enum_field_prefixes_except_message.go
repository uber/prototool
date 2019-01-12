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
	"github.com/uber/prototool/internal/strs"
	"github.com/uber/prototool/internal/text"
)

var enumFieldPrefixesExceptMessageLinter = NewLinter(
	"ENUM_FIELD_PREFIXES_EXCEPT_MESSAGE",
	"Verifies that all enum fields are prefixed with ENUM_NAME_.",
	checkEnumFieldPrefixesExceptMessage,
)

func checkEnumFieldPrefixesExceptMessage(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&enumFieldPrefixesExceptMessageVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type enumFieldPrefixesExceptMessageVisitor struct {
	baseAddVisitor
}

func (v *enumFieldPrefixesExceptMessageVisitor) VisitMessage(message *proto.Message) {
	for _, child := range message.Elements {
		child.Accept(v)
	}
}

func (v *enumFieldPrefixesExceptMessageVisitor) VisitEnum(enum *proto.Enum) {
	for _, child := range enum.Elements {
		child.Accept(v)
	}
}

func (v *enumFieldPrefixesExceptMessageVisitor) VisitEnumField(enumField *proto.EnumField) {
	parent := enumField.Parent
	if parent == nil {
		v.AddFailuref(enumField.Position, "System error. Enum field %q has no parent.", enumField.Name)
	}
	enum, ok := parent.(*proto.Enum)
	if !ok {
		v.AddFailuref(enumField.Position, "System error. Enum field %q has no enum parent.", enumField.Name)
	}
	expectedPrefix := strs.ToUpperSnakeCase(enum.Name) + "_"
	if !strings.HasPrefix(enumField.Name, expectedPrefix) {
		v.AddFailuref(enumField.Position, "Enum field %q is expected to have the prefix %q.", enumField.Name, expectedPrefix)
	}
}
