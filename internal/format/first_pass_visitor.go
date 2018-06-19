// Copyright (c) 2018 Uber Technologies, Inc.
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

package format

import (
	"sort"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/text"
)

var _ proto.Visitor = &firstPassVisitor{}

type firstPassVisitor struct {
	*baseVisitor

	Syntax  *proto.Syntax
	Package *proto.Package
	Options []*proto.Option
	Imports []*proto.Import

	haveHitNonComment bool
}

func newFirstPassVisitor(config settings.Config) *firstPassVisitor {
	return &firstPassVisitor{baseVisitor: newBaseVisitor(config.Format.Indent)}
}

func (v *firstPassVisitor) Do() []*text.Failure {
	if v.Syntax != nil {
		v.PComment(v.Syntax.Comment)
		if v.Syntax.Comment != nil {
			// special case
			v.P()
		}
		v.PWithInlineComment(v.Syntax.InlineComment, `syntax = "`, v.Syntax.Value, `";`)
		v.P()
	}
	if v.Package != nil {
		v.PComment(v.Package.Comment)
		v.PWithInlineComment(v.Package.InlineComment, `package `, v.Package.Name, `;`)
		v.P()
	}
	if len(v.Options) > 0 {
		v.POptions(false, v.Options...)
		v.P()
	}
	if len(v.Imports) > 0 {
		v.PImports(v.Imports)
		v.P()
	}
	return v.Failures
}

func (v *firstPassVisitor) VisitMessage(element *proto.Message) {
	v.haveHitNonComment = true
}

func (v *firstPassVisitor) VisitService(element *proto.Service) {
	v.haveHitNonComment = true
}

func (v *firstPassVisitor) VisitSyntax(element *proto.Syntax) {
	v.haveHitNonComment = true
	if v.Syntax != nil {
		v.AddFailure(element.Position, "duplicate syntax specified")
		return
	}
	v.Syntax = element
}

func (v *firstPassVisitor) VisitPackage(element *proto.Package) {
	v.haveHitNonComment = true
	if v.Package != nil {
		v.AddFailure(element.Position, "duplicate package specified")
		return
	}
	v.Package = element
}

func (v *firstPassVisitor) VisitOption(element *proto.Option) {
	// this will only hit file options since we don't do any
	// visiting of children in this visitor
	v.haveHitNonComment = true
	v.Options = append(v.Options, element)
}

func (v *firstPassVisitor) VisitImport(element *proto.Import) {
	v.haveHitNonComment = true
	v.Imports = append(v.Imports, element)
}

func (v *firstPassVisitor) VisitNormalField(element *proto.NormalField) {
	v.haveHitNonComment = true
}

func (v *firstPassVisitor) VisitEnumField(element *proto.EnumField) {
	v.haveHitNonComment = true
}

func (v *firstPassVisitor) VisitEnum(element *proto.Enum) {
	v.haveHitNonComment = true
}

func (v *firstPassVisitor) VisitComment(element *proto.Comment) {
	// We only print file-level comments before syntax, package, file-level options,
	// or package if they are at the top of the file
	if !v.haveHitNonComment {
		v.PComment(element)
		v.P()
	}
}

func (v *firstPassVisitor) VisitOneof(element *proto.Oneof) {
	v.haveHitNonComment = true
}

func (v *firstPassVisitor) VisitOneofField(element *proto.OneOfField) {
	v.haveHitNonComment = true
}

func (v *firstPassVisitor) VisitReserved(element *proto.Reserved) {
	v.haveHitNonComment = true
}

func (v *firstPassVisitor) VisitRPC(element *proto.RPC) {
	v.haveHitNonComment = true
}

func (v *firstPassVisitor) VisitMapField(element *proto.MapField) {
	v.haveHitNonComment = true
}

func (v *firstPassVisitor) VisitGroup(element *proto.Group) {
	v.haveHitNonComment = true
}

func (v *firstPassVisitor) VisitExtensions(element *proto.Extensions) {
	v.haveHitNonComment = true
}

func (v *firstPassVisitor) PImports(imports []*proto.Import) {
	if len(imports) == 0 {
		return
	}
	sort.Slice(imports, func(i int, j int) bool { return imports[i].Filename < imports[j].Filename })
	for _, i := range imports {
		v.PComment(i.Comment)
		// kind can be "weak", "public", or empty
		// if weak or public, just print it out but with a space afterwards
		// otherwise do not print anything
		// https://developers.google.com/protocol-buffers/docs/reference/proto3-spec#import_statement
		kind := i.Kind
		if kind != "" {
			kind = kind + " "
		}
		v.PWithInlineComment(i.InlineComment, `import `, kind, `"`, i.Filename, `";`)
	}
}
