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

var fieldsNotReservedLinter = NewLinter(
	"FIELDS_NOT_RESERVED",
	`Verifies that no message or enum has a reserved field.`,
	checkFieldsNotReserved,
)

func checkFieldsNotReserved(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(fieldsNotReservedVisitor{baseAddVisitor: newBaseAddVisitor(add), outerNameStack: make([]string, 0)}, descriptors)
}

type fieldsNotReservedVisitor struct {
	baseAddVisitor
	outerNameStack []string
}

func (v fieldsNotReservedVisitor) VisitMessage(message *proto.Message) {
	// Push name on stack.
	v.outerNameStack = append(
		v.outerNameStack,
		message.Name,
	)
	for _, element := range message.Elements {
		element.Accept(v)
	}
	// Pop name after visiting.
	v.outerNameStack = v.outerNameStack[:len(v.outerNameStack)-1]
}

func (v fieldsNotReservedVisitor) VisitEnum(enum *proto.Enum) {
	// Push name on stack.
	v.outerNameStack = append(
		v.outerNameStack,
		enum.Name,
	)
	for _, element := range enum.Elements {
		element.Accept(v)
	}
	// Pop after visiting.
	v.outerNameStack = v.outerNameStack[:len(v.outerNameStack)-1]
}

func (v fieldsNotReservedVisitor) VisitReserved(reserved *proto.Reserved) {
	var reservedType string
	// A reserved can't have both ranges and field names in the same
	// field.
	if len(reserved.Ranges) > 0 {
		reservedType = "field numbers"
	} else {
		reservedType = "field names"
	}
	v.AddFailuref(reserved.Position, `%q cannot have reserved %s, use the deprecated field option instead.`, strings.Join(v.outerNameStack, "."), reservedType)
}
