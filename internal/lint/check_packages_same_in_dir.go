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
	"sort"
	"strings"
	"text/scanner"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

var packagesSameInDirLinter = NewLinter(
	"PACKAGES_SAME_IN_DIR",
	"Verifies that the packages of all files in a directory are the same.",
	checkPackagesSameInDir,
)

func checkPackagesSameInDir(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	visitor := &packagesSameInDirVisitor{baseAddVisitor: newBaseAddVisitor(add), pkgs: make(map[string]struct{})}
	if err := runVisitor(visitor, descriptors); err != nil {
		return err
	}
	if len(visitor.pkgs) > 1 {
		pkgsSlice := make([]string, 0, len(visitor.pkgs))
		for pkg := range visitor.pkgs {
			pkgsSlice = append(pkgsSlice, pkg)
		}
		sort.Strings(pkgsSlice)
		for _, descriptor := range descriptors {
			visitor.AddFailuref(scanner.Position{Filename: descriptor.Filename}, "Multiple packages in directory: %v.", strings.Join(pkgsSlice, ", "))
		}
	}
	return nil
}

type packagesSameInDirVisitor struct {
	baseAddVisitor

	pkgs map[string]struct{}

	pkg *proto.Package
}

func (v *packagesSameInDirVisitor) OnStart(*FileDescriptor) error {
	v.pkg = nil
	return nil
}

func (v *packagesSameInDirVisitor) VisitPackage(pkg *proto.Package) {
	if v.pkg != nil {
		v.AddFailuref(pkg.Position, "multiple package declarations, first was %v", v.pkg)
		return
	}
	v.pkg = pkg
}

func (v *packagesSameInDirVisitor) Finally() error {
	if v.pkg != nil {
		v.pkgs[v.pkg.Name] = struct{}{}
	}
	return nil
}
