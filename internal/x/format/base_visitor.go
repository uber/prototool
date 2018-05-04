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
	"fmt"
	"sort"
	"strings"
	"text/scanner"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/x/text"
)

type baseVisitor struct {
	*printer

	Failures []*text.Failure
}

func newBaseVisitor(indent string) *baseVisitor {
	return &baseVisitor{printer: newPrinter(indent)}
}

//func (v *baseVisitor) VisitMessage(element *proto.Message)         {}
//func (v *baseVisitor) VisitService(element *proto.Service)         {}
//func (v *baseVisitor) VisitSyntax(element *proto.Syntax)           {}
//func (v *baseVisitor) VisitPackage(element *proto.Package)         {}
//func (v *baseVisitor) VisitOption(element *proto.Option)           {}
//func (v *baseVisitor) VisitImport(element *proto.Import)           {}
//func (v *baseVisitor) VisitNormalField(element *proto.NormalField) {}
//func (v *baseVisitor) VisitEnumField(element *proto.EnumField)     {}
//func (v *baseVisitor) VisitEnum(element *proto.Enum)               {}
//func (v *baseVisitor) VisitComment(element *proto.Comment)         {}
//func (v *baseVisitor) VisitOneof(element *proto.Oneof)             {}
//func (v *baseVisitor) VisitOneofField(element *proto.OneOfField)   {}
//func (v *baseVisitor) VisitReserved(element *proto.Reserved)       {}
//func (v *baseVisitor) VisitRPC(element *proto.RPC)                 {}
//func (v *baseVisitor) VisitMapField(element *proto.MapField)       {}
//func (v *baseVisitor) VisitGroup(element *proto.Group)             {}
//func (v *baseVisitor) VisitExtensions(element *proto.Extensions)   {}

func (v *baseVisitor) AddFailure(position scanner.Position, format string, args ...interface{}) {
	v.Failures = append(v.Failures, &text.Failure{
		Line:    position.Line,
		Column:  position.Column,
		Message: fmt.Sprintf(format, args...),
	})
}

func (v *baseVisitor) PWithInlineComment(inlineComment *proto.Comment, args ...interface{}) {
	if inlineComment == nil || len(inlineComment.Lines) == 0 {
		v.P(args...)
		return
	}
	// https://github.com/emicklei/proto/commit/5a91db7561a4dedab311f36304fcf0512343a9b1
	//if inlineComment.Cstyle {
	//args = append(args, inlineComment.Lines[0])
	//v.P(args...)
	//for _, line := range inlineComment.Lines[1:] {
	//v.P(line)
	//}
	//return
	//}
	args = append(args, ` //`, cleanCommentLine(inlineComment.Lines[0]))
	v.P(args...)
	for i, line := range inlineComment.Lines[1:] {
		line = cleanCommentLine(line)
		if line == "" && i != len(inlineComment.Lines)-1 {
			v.P(`//`)
		} else {
			v.P(`//`, line)
		}
	}
}

func (v *baseVisitor) PComment(comment *proto.Comment) {
	if comment == nil || len(comment.Lines) == 0 {
		return
	}
	// https://github.com/emicklei/proto/commit/5a91db7561a4dedab311f36304fcf0512343a9b1
	//if comment.Cstyle {
	//for _, line := range comment.Lines {
	//v.P(line)
	//}
	//return
	//}
	// this is weird for now
	// we always want non-c-style after formatting
	for i, line := range comment.Lines {
		line = cleanCommentLine(line)
		if line == "" && !(i == 0 || i == len(comment.Lines)-1) {
			v.P(`//`)
		} else {
			v.P(`//`, line)
		}
	}
}

func (v *baseVisitor) POptions(isFieldOption bool, options ...*proto.Option) {
	if len(options) == 0 {
		return
	}
	sort.Slice(options, func(i int, j int) bool { return options[i].Name < options[j].Name })
	prefix := "option "
	if isFieldOption {
		prefix = ""
	}
	for i, o := range options {
		suffix := ";"
		if isFieldOption {
			if len(options) > 1 && i != len(options)-1 {
				suffix = ","
			} else {
				suffix = ""
			}
		}
		v.PComment(o.Comment)
		// TODO: this is a good example of the reasoning for https://github.com/uber/prototool/issues/1
		if len(o.Constant.Array) == 0 && len(o.Constant.OrderedMap) == 0 {
			v.PWithInlineComment(o.InlineComment, prefix, o.Name, ` = `, o.Constant.SourceRepresentation(), suffix)
		} else if len(o.Constant.Array) > 0 { // both Array and OrderedMap should not be set simultaneously, need more followup with emicklei/proto
			// TODO
		} else { // len(o.Constant.OrderedMap) > 0
			v.P(prefix, o.Name, ` = {`)
			v.In()
			for _, namedLiteral := range o.Constant.OrderedMap {
				v.P(namedLiteral.Name, `: `, namedLiteral.SourceRepresentation())
			}
			v.Out()
			v.PWithInlineComment(o.InlineComment, `}`, suffix)
		}
	}
}

func (v *baseVisitor) PField(prefix string, t string, field *proto.Field) {
	v.PComment(field.Comment)
	if len(field.Options) == 0 {
		v.PWithInlineComment(field.InlineComment, prefix, t, " ", field.Name, " = ", field.Sequence, ";")
		return
	}
	v.P(prefix, t, " ", field.Name, " = ", field.Sequence, " [")
	v.In()
	v.POptions(true, field.Options...)
	v.Out()
	v.PWithInlineComment(field.InlineComment, "];")
}

func cleanCommentLine(line string) string {
	// TODO: this is not great
	return strings.TrimLeft(line, "/")
}
