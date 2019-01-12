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

var fileOptionsCSharpNamespaceSameInDirLinter = NewLinter(
	"FILE_OPTIONS_CSHARP_NAMESPACE_SAME_IN_DIR",
	`Verifies that the file option "csharp_namespace" of all files in a directory are the same.`,
	newCheckFileOptionsSameInDir("csharp_namespace"),
)

var fileOptionsGoPackageSameInDirLinter = NewLinter(
	"FILE_OPTIONS_GO_PACKAGE_SAME_IN_DIR",
	`Verifies that the file option "go_package" of all files in a directory are the same.`,
	newCheckFileOptionsSameInDir("go_package"),
)

var fileOptionsJavaMultipleFilesSameInDirLinter = NewLinter(
	"FILE_OPTIONS_JAVA_MULTIPLE_FILES_SAME_IN_DIR",
	`Verifies that the file option "java_multiple_files" of all files in a directory are the same.`,
	newCheckFileOptionsSameInDir("java_multiple_files"),
)

var fileOptionsJavaPackageSameInDirLinter = NewLinter(
	"FILE_OPTIONS_JAVA_PACKAGE_SAME_IN_DIR",
	`Verifies that the file option "java_package" of all files in a directory are the same.`,
	newCheckFileOptionsSameInDir("java_package"),
)

var fileOptionsOBJCClassPrefixSameInDirLinter = NewLinter(
	"FILE_OPTIONS_OBJC_CLASS_PREFIX_SAME_IN_DIR",
	`Verifies that the file option "objc_class_prefix" of all files in a directory are the same.`,
	newCheckFileOptionsSameInDir("objc_class_prefix"),
)

var fileOptionsPHPNamespaceSameInDirLinter = NewLinter(
	"FILE_OPTIONS_PHP_NAMESPACE_SAME_IN_DIR",
	`Verifies that the file option "php_namespace" of all files in a directory are the same.`,
	newCheckFileOptionsSameInDir("php_namespace"),
)

func newCheckFileOptionsSameInDir(fileOption string) func(func(*text.Failure), string, []*FileDescriptor) error {
	return func(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
		visitor := &fileOptionsSameInDirVisitor{
			baseAddVisitor:   newBaseAddVisitor(add),
			fileOption:       fileOption,
			fileOptionValues: make(map[string]struct{}),
		}
		if err := runVisitor(visitor, descriptors); err != nil {
			return err
		}
		if len(visitor.fileOptionValues) > 1 {
			if _, ok := visitor.fileOptionValues[""]; ok {
				for _, descriptor := range descriptors {
					visitor.AddFailuref(scanner.Position{Filename: descriptor.Filename}, "File option %q set in some files in directory but not in others.", fileOption)
				}
				return nil
			}
			fileOptionValuesSlice := make([]string, 0, len(visitor.fileOptionValues))
			for fileOptionValue := range visitor.fileOptionValues {
				fileOptionValuesSlice = append(fileOptionValuesSlice, fileOptionValue)
			}
			sort.Strings(fileOptionValuesSlice)
			for _, descriptor := range descriptors {
				visitor.AddFailuref(scanner.Position{Filename: descriptor.Filename}, "Multiple values for file option %q in directory: %v.", fileOption, strings.Join(fileOptionValuesSlice, ", "))
			}
		}
		return nil
	}
}

type fileOptionsSameInDirVisitor struct {
	baseAddVisitor

	fileOption string

	fileOptionValues map[string]struct{}

	option *proto.Option
}

func (v *fileOptionsSameInDirVisitor) OnStart(*FileDescriptor) error {
	v.option = nil
	return nil
}

func (v *fileOptionsSameInDirVisitor) VisitOption(element *proto.Option) {
	// TODO: not validating this is a file option, or are we since we're not recursing on other elements?
	if element.Name == v.fileOption {
		if v.option != nil {
			v.AddFailuref(element.Position, "multiple option declarations for %s, first was %v", v.fileOption, v.option)
			return
		}
		v.option = element
	}
}

func (v *fileOptionsSameInDirVisitor) Finally() error {
	value := ""
	if v.option != nil {
		value = v.option.Constant.Source
	}
	v.fileOptionValues[value] = struct{}{}
	return nil
}
