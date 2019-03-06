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
	"github.com/uber/prototool/internal/text"
)

var fileOptionsUnsetJavaMultipleFilesLinter = NewLinter(
	"FILE_OPTIONS_UNSET_JAVA_MULTIPLE_FILES",
	`Verifies that the file option "java_multiple_files" is unset.`,
	newCheckFileOptionsUnset("java_multiple_files"),
)

var fileOptionsUnsetJavaOuterClassnameLinter = NewLinter(
	"FILE_OPTIONS_UNSET_JAVA_OUTER_CLASSNAME",
	`Verifies that the file option "java_outer_classname" is unset.`,
	newCheckFileOptionsUnset("java_outer_classname"),
)

func newCheckFileOptionsUnset(fileOption string) func(func(*text.Failure), string, []*FileDescriptor) error {
	return func(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
		return runVisitor(&fileOptionsUnsetVisitor{
			baseAddVisitor: newBaseAddVisitor(add),
			fileOption:     fileOption,
		}, descriptors)
	}
}

type fileOptionsUnsetVisitor struct {
	baseAddVisitor

	fileOption string

	seen     bool
	position scanner.Position
}

func (v *fileOptionsUnsetVisitor) OnStart(descriptor *FileDescriptor) error {
	v.seen = false
	v.position = scanner.Position{}
	return nil
}

func (v *fileOptionsUnsetVisitor) VisitOption(element *proto.Option) {
	// since we are not recursing on any elements, this is a file option
	if element.Name == v.fileOption {
		v.seen = true
		v.position = element.Position
	}
}

func (v *fileOptionsUnsetVisitor) Finally() error {
	if v.seen {
		v.AddFailuref(v.position, "File option %q should not be set.", v.fileOption)
	}
	return nil
}
