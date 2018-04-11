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
	"strings"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/x/text"
)

var fileOptionsEqualGoPackagePbSuffixChecker = NewAddChecker(
	"FILE_OPTIONS_EQUAL_GO_PACKAGE_PB_SUFFIX",
	`Verifies that the file option "go_package" is equal to $(basename PACKAGE)pb.`,
	newCheckFileOptionsEqual("go_package", func(pkg *proto.Package) string {
		return packageBasename(pkg.Name) + "pb"
	}),
)

var fileOptionsEqualJavaMultipleFilesTrueChecker = NewAddChecker(
	"FILE_OPTIONS_EQUAL_JAVA_MULTIPLE_FILES_TRUE",
	`Verifies that the file option "java_multiple_files" is equal to true.`,
	newCheckFileOptionsEqual("java_multiple_files", func(pkg *proto.Package) string {
		return "true"
	}),
)

var fileOptionsEqualJavaOuterClassnameProtoSuffixChecker = NewAddChecker(
	"FILE_OPTIONS_EQUAL_JAVA_OUTER_CLASSNAME_PROTO_SUFFIX",
	`Verifies that the file option "java_outer_classname" is equal to $(capitalize $(basename PACKAGE))Proto.`,
	newCheckFileOptionsEqual("java_outer_classname", func(pkg *proto.Package) string {
		// TODO: make sure strings.Title does what you want
		return strings.Title(packageBasename(pkg.Name)) + "Proto"
	}),
)

var fileOptionsEqualJavaPackageComPbChecker = NewAddChecker(
	"FILE_OPTIONS_EQUAL_JAVA_PACKAGE_COM_PB",
	`Verifies that the file option "java_package" is equal to com.PACKAGE.pb.`,
	newCheckFileOptionsEqual("java_package", func(pkg *proto.Package) string {
		return "com." + pkg.Name + ".pb"
	}),
)

func newCheckFileOptionsEqual(fileOption string, expectedValueFunc func(*proto.Package) string) func(func(*text.Failure), string, []*proto.Proto) error {
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
	expectedValueFunc func(*proto.Package) string

	pkg    *proto.Package
	option *proto.Option
}

func (v *fileOptionsEqualVisitor) OnStart(*proto.Proto) error {
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
	if v.pkg == nil || v.option == nil {
		// do not do anything, other linters should verify that the file option exists
		// this makes it possible to be optional if a required file option linter is suppressed
		// TODO make sure this is consistent across all linters
		return nil
	}
	// TODO: handle AggregatedConstants
	value := v.option.Constant.Source
	expectedValue := v.expectedValueFunc(v.pkg)
	if expectedValue != value {
		v.AddFailuref(v.option.Position, "Expected %q for option %q but was %q.", expectedValue, v.option.Name, value)
	}
	return nil
}

func packageBasename(pkg string) string {
	split := strings.Split(pkg, ".")
	return split[len(split)-1]
}
