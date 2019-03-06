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
	"text/scanner"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

var syntaxProto3Linter = NewLinter(
	"SYNTAX_PROTO3",
	"Verifies that the syntax is proto3.",
	checkSyntaxProto3,
)

func checkSyntaxProto3(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&syntaxProto3Visitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type syntaxProto3Visitor struct {
	baseAddVisitor

	filename string
	syntax   *proto.Syntax
}

func (v *syntaxProto3Visitor) OnStart(descriptor *FileDescriptor) error {
	v.filename = descriptor.Filename
	v.syntax = nil
	return nil
}

func (v *syntaxProto3Visitor) VisitSyntax(syntax *proto.Syntax) {
	if v.syntax != nil {
		v.AddFailuref(syntax.Position, "multiple syntax declarations, first was %v", v.syntax)
		return
	}
	v.syntax = syntax
}

func (v *syntaxProto3Visitor) Finally() error {
	if v.syntax == nil {
		v.AddFailuref(scanner.Position{Filename: v.filename}, "No syntax declaration found.")
		return nil
	}
	if v.syntax.Value != "proto3" {
		v.AddFailuref(v.syntax.Position, "Syntax should be proto3 but was %q.", v.syntax.Value)
	}
	return nil
}
