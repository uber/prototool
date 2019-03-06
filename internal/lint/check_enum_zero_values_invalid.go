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

var enumZeroValuesInvalidLinter = NewLinter(
	"ENUM_ZERO_VALUES_INVALID",
	"Verifies that all enum zero value names are [NESTED_MESSAGE_NAME_]ENUM_NAME_INVALID.",
	checkEnumZeroValuesInvalid,
)

func checkEnumZeroValuesInvalid(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&enumZeroValuesInvalidVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type enumZeroValuesInvalidVisitor struct {
	baseAddVisitor

	nestedNames []string
}

func (v *enumZeroValuesInvalidVisitor) VisitMessage(message *proto.Message) {
	v.nestedNames = append(v.nestedNames, strs.ToUpperSnakeCase(message.Name))
	for _, child := range message.Elements {
		child.Accept(v)
	}
	v.nestedNames = v.nestedNames[0 : len(v.nestedNames)-1]
}

func (v *enumZeroValuesInvalidVisitor) VisitEnum(enum *proto.Enum) {
	v.nestedNames = append(v.nestedNames, strs.ToUpperSnakeCase(enum.Name))
	for _, child := range enum.Elements {
		child.Accept(v)
	}
	v.nestedNames = v.nestedNames[0 : len(v.nestedNames)-1]
}

func (v *enumZeroValuesInvalidVisitor) VisitEnumField(enumField *proto.EnumField) {
	if enumField.Integer == 0 {
		expectedName := strings.Join(v.nestedNames, "_") + "_INVALID"
		if enumField.Name != expectedName {
			v.AddFailuref(enumField.Position, "Zero value enum field %q is expected to have the name %q.", enumField.Name, expectedName)
		}
	}
}
