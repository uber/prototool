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
	"strings"
	"text/scanner"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

const namesSuppressableAnnotation = "naming"

var (
	namesNoCommonLinter = newNamesLinter(
		"common",
		"Common has no semantic meaning, consider using a name that reflects the type instead.",
	)
	namesNoDataLinter = newNamesLinter(
		"data",
		`Data is a decorator and all types on Protobuf are data, consider merging this information into a higher-level type, or if you must have such a type, Use "Info" instead.`,
	)
	namesNoUUIDLinter = newNamesLinter(
		"uuid",
		`UUIDs in Protobuf are named ID instead of UUID.`,
	)
)

func newNamesLinter(outlawedName string, additionalHelp string) Linter {
	return NewLinter(
		"NAMES_NO_"+strings.ToUpper(outlawedName),
		fmt.Sprintf(
			`Suppressable with "@suppresswarnings %s". Verifies that no type name contains the word %q.`,
			namesSuppressableAnnotation,
			outlawedName,
		),
		newCheckNames(
			outlawedName,
			additionalHelp,
		),
	)
}

func newCheckNames(outlawedName string, additionalHelp string) func(func(*text.Failure), string, []*FileDescriptor) error {
	return func(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
		return runVisitor(&namesVisitor{
			baseAddVisitor: newBaseAddVisitor(add),
			outlawedName:   outlawedName,
			additionalHelp: additionalHelp,
		}, descriptors)
	}
}

type namesVisitor struct {
	baseAddVisitor

	outlawedName   string
	additionalHelp string
}

func (v *namesVisitor) VisitMessage(element *proto.Message) {
	v.checkName(element.Position, element.Name)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v *namesVisitor) VisitService(element *proto.Service) {
	v.checkName(element.Position, element.Name)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v *namesVisitor) VisitSyntax(element *proto.Syntax) {
	// do nothing
}

func (v *namesVisitor) VisitPackage(element *proto.Package) {
	v.checkName(element.Position, element.Name)
}

func (v *namesVisitor) VisitOption(element *proto.Option) {
	v.checkName(element.Position, element.Name)
}

func (v *namesVisitor) VisitImport(element *proto.Import) {
	// do nothing
}

func (v *namesVisitor) VisitNormalField(element *proto.NormalField) {
	v.checkName(element.Position, element.Name)
	for _, child := range element.Options {
		child.Accept(v)
	}
}

func (v *namesVisitor) VisitEnumField(element *proto.EnumField) {
	v.checkName(element.Position, element.Name)
	if element.ValueOption != nil {
		element.ValueOption.Accept(v)
	}
}

func (v *namesVisitor) VisitEnum(element *proto.Enum) {
	v.checkName(element.Position, element.Name)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v *namesVisitor) VisitComment(element *proto.Comment) {
	// do nothing
}

func (v *namesVisitor) VisitOneof(element *proto.Oneof) {
	v.checkName(element.Position, element.Name)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v *namesVisitor) VisitOneofField(element *proto.OneOfField) {
	v.checkName(element.Position, element.Name)
	for _, child := range element.Options {
		child.Accept(v)
	}
}

func (v *namesVisitor) VisitReserved(element *proto.Reserved) {
	for _, fieldName := range element.FieldNames {
		v.checkName(element.Position, fieldName)
	}
}

func (v *namesVisitor) VisitRPC(element *proto.RPC) {
	v.checkName(element.Position, element.Name)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v *namesVisitor) VisitMapField(element *proto.MapField) {
	v.checkName(element.Position, element.Name)
	for _, child := range element.Options {
		child.Accept(v)
	}
}

func (v *namesVisitor) VisitGroup(element *proto.Group) {
	v.checkName(element.Position, element.Name)
	for _, child := range element.Elements {
		child.Accept(v)
	}
}

func (v *namesVisitor) VisitExtensions(element *proto.Extensions) {
	// do nothing
}

func (v *namesVisitor) checkName(position scanner.Position, name string) {
	if strings.Contains(strings.ToLower(name), v.outlawedName) {
		v.AddFailuref(position, `The name %q contains the outlawed name %q. %s This can be suppressed by adding "@suppresswarnings %s" to the type comment.`, name, v.outlawedName, v.additionalHelp, namesSuppressableAnnotation)
	}
}
