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
	"github.com/uber/prototool/internal/protostrs"
	"github.com/uber/prototool/internal/text"
)

var packageMajorVersionedLinter = NewLinter(
	"PACKAGE_MAJOR_VERSIONED",
	`Verifies that the package is of the form "package.vMAJORVERSION".`,
	checkPackageMajorVersioned,
)

func checkPackageMajorVersioned(add func(*text.Failure), dirPath string, descriptors []*proto.Proto) error {
	return runVisitor(&packageMajorVersionedVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type packageMajorVersionedVisitor struct {
	baseAddVisitor

	filename string
	pkg      *proto.Package
}

func (v *packageMajorVersionedVisitor) OnStart(descriptor *proto.Proto) error {
	v.filename = descriptor.Filename
	v.pkg = nil
	return nil
}

func (v *packageMajorVersionedVisitor) VisitPackage(pkg *proto.Package) {
	if v.pkg != nil {
		v.AddFailuref(pkg.Position, "multiple package declarations, first was %v", v.pkg)
		return
	}
	v.pkg = pkg
}

func (v *packageMajorVersionedVisitor) Finally() error {
	if v.pkg == nil {
		v.AddFailuref(scanner.Position{Filename: v.filename}, "No package declaration found.")
		return nil
	}
	if _, ok := protostrs.MajorVersion(v.pkg.Name); !ok {
		v.AddFailuref(v.pkg.Position, `Package should be of the form "package.vMAJORVERSION" but was %q.`, v.pkg.Name)
	}
	return nil
}
