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

package lint

import (
	"text/scanner"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/failure"
)

type baseVisitor struct{}

func (baseVisitor) OnStart(*proto.Proto) error { return nil }
func (baseVisitor) Finally() error             { return nil }

func (baseVisitor) VisitMessage(m *proto.Message)         {}
func (baseVisitor) VisitService(s *proto.Service)         {}
func (baseVisitor) VisitSyntax(s *proto.Syntax)           {}
func (baseVisitor) VisitPackage(p *proto.Package)         {}
func (baseVisitor) VisitOption(o *proto.Option)           {}
func (baseVisitor) VisitImport(i *proto.Import)           {}
func (baseVisitor) VisitNormalField(i *proto.NormalField) {}
func (baseVisitor) VisitEnumField(i *proto.EnumField)     {}
func (baseVisitor) VisitEnum(e *proto.Enum)               {}
func (baseVisitor) VisitComment(e *proto.Comment)         {}
func (baseVisitor) VisitOneof(o *proto.Oneof)             {}
func (baseVisitor) VisitOneofField(o *proto.OneOfField)   {}
func (baseVisitor) VisitReserved(r *proto.Reserved)       {}
func (baseVisitor) VisitRPC(r *proto.RPC)                 {}
func (baseVisitor) VisitMapField(f *proto.MapField)       {}
func (baseVisitor) VisitGroup(g *proto.Group)             {}
func (baseVisitor) VisitExtensions(e *proto.Extensions)   {}

type baseAddVisitor struct {
	baseVisitor
	add func(*failure.Failure)
}

func newBaseAddVisitor(add func(*failure.Failure)) baseAddVisitor {
	return baseAddVisitor{add: add}
}

func (v baseAddVisitor) AddFailuref(position scanner.Position, format string, args ...interface{}) {
	v.add(failure.Newf(position, failure.Lint, format, args...))
}

// extendedVisitor extends the proto.Visitor interface.
// extendedVisitors are expected to be called with one file at a time,
// and are not thread-safe.
type extendedVisitor interface {
	proto.Visitor

	// OnStart is called when visiting is started.
	OnStart(*proto.Proto) error
	// Finally is called when visiting is done.
	Finally() error
}

func runVisitor(visitor extendedVisitor, descriptors []*proto.Proto) error {
	for _, descriptor := range descriptors {
		if err := visitor.OnStart(descriptor); err != nil {
			return err
		}
		for _, element := range descriptor.Elements {
			element.Accept(visitor)
		}
		if err := visitor.Finally(); err != nil {
			return err
		}
	}
	return nil
}
