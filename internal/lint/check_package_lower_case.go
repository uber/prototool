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

var packageLowerCaseLinter = NewLinter(
	"PACKAGE_LOWER_CASE",
	"Verifies that the package name only contains characters in the range a-z0-9 and periods.",
	checkPackageLowerCase,
)

func checkPackageLowerCase(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&packageLowerCaseVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type packageLowerCaseVisitor struct {
	baseAddVisitor

	filename string
	pkg      *proto.Package
}

func (v *packageLowerCaseVisitor) OnStart(descriptor *FileDescriptor) error {
	v.filename = descriptor.Filename
	v.pkg = nil
	return nil
}

func (v *packageLowerCaseVisitor) VisitPackage(pkg *proto.Package) {
	if v.pkg != nil {
		v.AddFailuref(pkg.Position, "multiple package declarations, first was %v", v.pkg)
		return
	}
	v.pkg = pkg
}

func (v *packageLowerCaseVisitor) Finally() error {
	if v.pkg != nil {
		if !isPackageLowerCase(v.pkg.Name) {
			v.AddFailuref(v.pkg.Position, "Packages should only contain characters in the range a-z0-9 and periods but was %q.", v.pkg.Name)
		}
	}
	return nil
}

func isPackageLowerCase(packageName string) bool {
	for _, r := range packageName {
		if !(('a' <= r && r <= 'z') || ('0' <= r && r <= '9') || r == '.') {
			return false
		}
	}
	return true
}
