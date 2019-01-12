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
	"github.com/uber/prototool/internal/strs"
	"github.com/uber/prototool/internal/text"
)

var enumZeroValuesInvalidExceptMessageLinter = NewLinter(
	"ENUM_ZERO_VALUES_INVALID_EXCEPT_MESSAGE",
	"Verifies that all enum zero value names are ENUM_NAME_INVALID.",
	checkEnumZeroValuesInvalidExceptMessage,
)

func checkEnumZeroValuesInvalidExceptMessage(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&enumZeroValuesInvalidExceptMessageVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type enumZeroValuesInvalidExceptMessageVisitor struct {
	baseAddVisitor
}

func (v *enumZeroValuesInvalidExceptMessageVisitor) VisitMessage(message *proto.Message) {
	for _, child := range message.Elements {
		child.Accept(v)
	}
}

func (v *enumZeroValuesInvalidExceptMessageVisitor) VisitEnum(enum *proto.Enum) {
	for _, child := range enum.Elements {
		child.Accept(v)
	}
}

func (v *enumZeroValuesInvalidExceptMessageVisitor) VisitEnumField(enumField *proto.EnumField) {
	if enumField.Integer == 0 {
		parent := enumField.Parent
		if parent == nil {
			v.AddFailuref(enumField.Position, "System error. Enum field %q has no parent.", enumField.Name)
		}
		enum, ok := parent.(*proto.Enum)
		if !ok {
			v.AddFailuref(enumField.Position, "System error. Enum field %q has no enum parent.", enumField.Name)
		}
		expectedName := strs.ToUpperSnakeCase(enum.Name) + "_INVALID"
		if enumField.Name != expectedName {
			v.AddFailuref(enumField.Position, "Zero value enum field %q is expected to have the name %q.", enumField.Name, expectedName)
		}
	}
}
