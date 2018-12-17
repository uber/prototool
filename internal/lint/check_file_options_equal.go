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
	"fmt"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/protostrs"
	"github.com/uber/prototool/internal/text"
)

var fileOptionsEqualGoPackagePbSuffixLinter = NewLinter(
	"FILE_OPTIONS_EQUAL_GO_PACKAGE_PB_SUFFIX",
	`Verifies that the file option "go_package" is equal to $(basename PACKAGE)pb.`,
	newCheckFileOptionsEqual("go_package", func(_ *proto.Proto, pkg *proto.Package) string {
		return protostrs.GoPackage(pkg.Name)
	}),
)

var fileOptionsEqualGoPackageLastTwoSuffixLinter = NewLinter(
	"FILE_OPTIONS_EQUAL_GO_PACKAGE_LAST_TWO_SUFFIX",
	`Verifies that the file option "go_package" is equal to the last two values of the package separated by "."s, or just the package name if there are no "."s.`,
	newCheckFileOptionsEqual("go_package", func(_ *proto.Proto, pkg *proto.Package) string {
		return protostrs.GoPackageLastTwo(pkg.Name)
	}),
)

var fileOptionsEqualJavaMultipleFilesTrueLinter = NewLinter(
	"FILE_OPTIONS_EQUAL_JAVA_MULTIPLE_FILES_TRUE",
	`Verifies that the file option "java_multiple_files" is equal to true.`,
	newCheckFileOptionsEqual("java_multiple_files", func(*proto.Proto, *proto.Package) string {
		return "true"
	}),
)

var fileOptionsEqualJavaOuterClassnameProtoSuffixLinter = NewLinter(
	"FILE_OPTIONS_EQUAL_JAVA_OUTER_CLASSNAME_PROTO_SUFFIX",
	`Verifies that the file option "java_outer_classname" is equal to $(upperCamelCase $(basename FILE))Proto.`,
	newCheckFileOptionsEqual("java_outer_classname", func(descriptor *proto.Proto, _ *proto.Package) string {
		return protostrs.JavaOuterClassname(descriptor.Filename)
	}),
)

var fileOptionsEqualJavaPackageComPrefixLinter = NewLinter(
	"FILE_OPTIONS_EQUAL_JAVA_PACKAGE_COM_PREFIX",
	`Verifies that the file option "java_package" is equal to com.PACKAGE.`,
	newCheckFileOptionsEqual("java_package", func(_ *proto.Proto, pkg *proto.Package) string {
		return protostrs.JavaPackage(pkg.Name)
	}),
)

func newCheckFileOptionsEqual(fileOption string, expectedValueFunc func(*proto.Proto, *proto.Package) string) func(func(*text.Failure), string, []*proto.Proto) error {
	return func(add func(*text.Failure), dirPath string, descriptors []*proto.Proto) error {
		return runVisitor(&fileOptionsEqualVisitor{
			baseAddVisitor:    newBaseAddVisitor(add),
			fileOption:        fileOption,
			expectedValueFunc: expectedValueFunc,
		}, descriptors)
	}
}

type fileOptionsEqualVisitor struct {
	baseAddVisitor

	fileOption        string
	expectedValueFunc func(*proto.Proto, *proto.Package) string

	descriptor *proto.Proto
	pkg        *proto.Package
	option     *proto.Option
}

func (v *fileOptionsEqualVisitor) OnStart(descriptor *proto.Proto) error {
	v.descriptor = descriptor
	v.pkg = nil
	v.option = nil
	return nil
}

func (v *fileOptionsEqualVisitor) VisitPackage(element *proto.Package) {
	v.pkg = element
}

func (v *fileOptionsEqualVisitor) VisitOption(element *proto.Option) {
	// TODO: not validating this is a file option, or are we since we're not recursing on other elements?
	if element.Name == v.fileOption {
		v.option = element
	}
}

func (v *fileOptionsEqualVisitor) Finally() error {
	if v.descriptor == nil || v.pkg == nil || v.option == nil {
		// do not do anything, other linters should verify that the file option exists
		// this makes it possible to be optional if a required file option linter is suppressed
		// TODO make sure this is consistent across all linters
		return nil
	}
	if v.descriptor.Filename == "" {
		// if this isn't set, we made a mistake setting this up, return a system error
		return fmt.Errorf("expected filename to be set for descriptor %v in checkFileOptionsEqual linter", v.descriptor)
	}
	// TODO: handle AggregatedConstants
	value := v.option.Constant.Source
	expectedValue := v.expectedValueFunc(v.descriptor, v.pkg)
	if expectedValue != value {
		v.AddFailuref(v.option.Position, "Expected %q for option %q but was %q.", expectedValue, v.option.Name, value)
	}
	return nil
}
