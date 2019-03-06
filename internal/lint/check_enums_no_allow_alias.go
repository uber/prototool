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

var enumsNoAllowAliasLinter = NewLinter(
	"ENUMS_NO_ALLOW_ALIAS",
	`Verifies that no enums use the option "allow_alias".`,
	checkEnumsNoAllowAlias,
)

func checkEnumsNoAllowAlias(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(enumsNoAllowAliasVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type enumsNoAllowAliasVisitor struct {
	baseAddVisitor
}

func (v enumsNoAllowAliasVisitor) VisitMessage(element *proto.Message) {
	// for nested enums
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v enumsNoAllowAliasVisitor) VisitEnum(element *proto.Enum) {
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v enumsNoAllowAliasVisitor) VisitOption(element *proto.Option) {
	// file option
	if element.Parent == nil {
		return
	}
	switch element.Parent.(type) {
	case (*proto.Enum):
		if element.Name == "allow_alias" {
			v.AddFailuref(element.Position, "Enum aliases are not allowed.")
		}
	}
}
