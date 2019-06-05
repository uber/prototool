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

package format

import (
	"sort"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/protostrs"
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

	filename                 string
	fix                      int
	fileHeader               string
	javaPackagePrefix        string
	goPackageOption          *proto.Option
	javaMultipleFilesOption  *proto.Option
	javaOuterClassnameOption *proto.Option
	javaPackageOption        *proto.Option
	csharpNamespaceOption    *proto.Option
	objcClassPrefixOption    *proto.Option
	phpNamespaceOption       *proto.Option
}

func newFirstPassVisitor(filename string, fix int, fileHeader string, javaPackagePrefix string) *firstPassVisitor {
	return &firstPassVisitor{baseVisitor: newBaseVisitor(), filename: filename, fix: fix, fileHeader: fileHeader, javaPackagePrefix: javaPackagePrefix}
}

func (v *firstPassVisitor) Do() []*text.Failure {
	if v.fix != FixNone && v.fileHeader != "" {
		v.P(v.fileHeader)
		v.P()
	}
	if v.Syntax != nil {
		v.PComment(v.Syntax.Comment)
		if v.Syntax.Comment != nil {
			// special case, we add a newline in between the first comment and syntax
			// to separate licenses, file descriptions, etc.
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
	if v.fix != FixNone && v.Package != nil {
		if v.goPackageOption == nil {
			v.goPackageOption = &proto.Option{Name: "go_package"}
		}
		if v.javaMultipleFilesOption == nil {
			v.javaMultipleFilesOption = &proto.Option{Name: "java_multiple_files"}
		}
		if v.javaOuterClassnameOption == nil {
			v.javaOuterClassnameOption = &proto.Option{Name: "java_outer_classname"}
		}
		if v.javaPackageOption == nil {
			v.javaPackageOption = &proto.Option{Name: "java_package"}
		}
		if v.fix == FixV2 {
			if v.csharpNamespaceOption == nil {
				v.csharpNamespaceOption = &proto.Option{Name: "csharp_namespace"}
			}
			if v.objcClassPrefixOption == nil {
				v.objcClassPrefixOption = &proto.Option{Name: "objc_class_prefix"}
			}
			if v.phpNamespaceOption == nil {
				v.phpNamespaceOption = &proto.Option{Name: "php_namespace"}
			}
		}
		if v.fix == FixV2 {
			v.goPackageOption.Constant = proto.Literal{
				Source:    protostrs.GoPackageV2(v.Package.Name),
				IsString:  true,
				QuoteRune: '"',
			}
		} else {
			v.goPackageOption.Constant = proto.Literal{
				Source:    protostrs.GoPackage(v.Package.Name),
				IsString:  true,
				QuoteRune: '"',
			}
		}
		v.javaMultipleFilesOption.Constant = proto.Literal{
			Source: "true",
		}
		v.javaOuterClassnameOption.Constant = proto.Literal{
			Source:    protostrs.JavaOuterClassname(v.filename),
			IsString:  true,
			QuoteRune: '"',
		}
		v.javaPackageOption.Constant = proto.Literal{
			Source:    protostrs.JavaPackagePrefixOverride(v.Package.Name, v.javaPackagePrefix),
			IsString:  true,
			QuoteRune: '"',
		}
		if v.fix == FixV2 {
			v.csharpNamespaceOption.Constant = proto.Literal{
				Source:    protostrs.CSharpNamespace(v.Package.Name),
				IsString:  true,
				QuoteRune: '"',
			}
			v.objcClassPrefixOption.Constant = proto.Literal{
				Source:    protostrs.OBJCClassPrefix(v.Package.Name),
				IsString:  true,
				QuoteRune: '"',
			}
			v.phpNamespaceOption.Constant = proto.Literal{
				Source:    protostrs.PHPNamespace(v.Package.Name),
				IsString:  true,
				QuoteRune: '"',
			}
		}
		v.Options = append(
			v.Options,
			v.goPackageOption,
			v.javaMultipleFilesOption,
			v.javaOuterClassnameOption,
			v.javaPackageOption,
		)
		if v.fix == FixV2 {
			v.Options = append(
				v.Options,
				v.csharpNamespaceOption,
				v.objcClassPrefixOption,
				v.phpNamespaceOption,
			)
		}
	}
	if len(v.Options) > 0 {
		v.POptions(v.Options...)
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
	if v.fix != FixNone {
		switch element.Name {
		case "csharp_namespace":
			v.csharpNamespaceOption = element
			return
		case "go_package":
			v.goPackageOption = element
			return
		case "java_multiple_files":
			v.javaMultipleFilesOption = element
			return
		case "java_outer_classname":
			v.javaOuterClassnameOption = element
			return
		case "java_package":
			v.javaPackageOption = element
			return
		case "objc_class_prefix":
			v.objcClassPrefixOption = element
			return
		case "php_namespace":
			v.phpNamespaceOption = element
			return
		}
	}
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
		if v.fix == FixNone || v.fileHeader == "" {
			v.PComment(element)
			v.P()
		}
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
