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
	"github.com/uber/prototool/internal/x/text"
)

var fileOptionsRequireGoPackageChecker = NewAddChecker(
	"FILE_OPTIONS_REQUIRE_GO_PACKAGE",
	`Verifies that the file option "go_package" is set.`,
	newCheckFileOptionsRequire("go_package"),
)

var fileOptionsRequireJavaMultipleFilesChecker = NewAddChecker(
	"FILE_OPTIONS_REQUIRE_JAVA_MULTIPLE_FILES",
	`Verifies that the file option "java_multiple_files" is set.`,
	newCheckFileOptionsRequire("java_multiple_files"),
)

var fileOptionsRequireJavaOuterClassnameChecker = NewAddChecker(
	"FILE_OPTIONS_REQUIRE_JAVA_OUTER_CLASSNAME",
	`Verifies that the file option "java_outer_classname" is set.`,
	newCheckFileOptionsRequire("java_outer_classname"),
)

var fileOptionsRequireJavaPackageChecker = NewAddChecker(
	"FILE_OPTIONS_REQUIRE_JAVA_PACKAGE",
	`Verifies that the file option "java_package" is set.`,
	newCheckFileOptionsRequire("java_package"),
)

func newCheckFileOptionsRequire(fileOption string) func(func(*text.Failure), string, []*proto.Proto) error {
	return func(add func(*text.Failure), dirPath string, descriptors []*proto.Proto) error {
		return runVisitor(&fileOptionsRequireVisitor{
			baseAddVisitor: newBaseAddVisitor(add),
			fileOption:     fileOption,
		}, descriptors)
	}
}

type fileOptionsRequireVisitor struct {
	baseAddVisitor

	fileOption string

	filename string
	seen     bool
}

func (v *fileOptionsRequireVisitor) OnStart(descriptor *proto.Proto) error {
	v.filename = descriptor.Filename
	v.seen = false
	return nil
}

func (v *fileOptionsRequireVisitor) VisitOption(element *proto.Option) {
	// TODO: not validating this is a file option, or are we since we're not recursing on other elements?
	if element.Name == v.fileOption {
		v.seen = true
	}
}

func (v *fileOptionsRequireVisitor) Finally() error {
	if !v.seen {
		v.AddFailuref(scanner.Position{Filename: v.filename}, "File option %q is required.", v.fileOption)
	}
	return nil
}
