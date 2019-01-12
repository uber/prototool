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
	"github.com/uber/prototool/internal/protostrs"
	"github.com/uber/prototool/internal/text"
)

var packageMajorBetaVersionedLinter = NewLinter(
	"PACKAGE_MAJOR_BETA_VERSIONED",
	`Verifies that the package is of the form "package.vMAJORVERSION" or "package.vMAJORVERSIONbetaBETAVERSION" with versions >=1.`,
	checkPackageMajorBetaVersioned,
)

func checkPackageMajorBetaVersioned(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&packageMajorBetaVersionedVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type packageMajorBetaVersionedVisitor struct {
	baseAddVisitor

	filename string
	pkg      *proto.Package
}

func (v *packageMajorBetaVersionedVisitor) OnStart(descriptor *FileDescriptor) error {
	v.filename = descriptor.Filename
	v.pkg = nil
	return nil
}

func (v *packageMajorBetaVersionedVisitor) VisitPackage(pkg *proto.Package) {
	if v.pkg != nil {
		v.AddFailuref(pkg.Position, "multiple package declarations, first was %v", v.pkg)
		return
	}
	v.pkg = pkg
}

func (v *packageMajorBetaVersionedVisitor) Finally() error {
	if v.pkg == nil {
		v.AddFailuref(scanner.Position{Filename: v.filename}, "No package declaration found.")
		return nil
	}
	if _, _, ok := protostrs.MajorBetaVersion(v.pkg.Name); !ok {
		v.AddFailuref(v.pkg.Position, `Package should be of the form "package.vMAJORVERSION" or "packagevMAJORVERSIONbetaBETAVERSION" with versions >=1 but was %q.`, v.pkg.Name)
	}
	return nil
}
