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
	"strings"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

var fileOptionsGoPackageNotLongFormLinter = NewLinter(
	"FILE_OPTIONS_GO_PACKAGE_NOT_LONG_FORM",
	`Verifies that the file option "go_package" is not of the form "go/import/path;package".`,
	checkFileOptionsGoPackageNotLongForm,
)

func checkFileOptionsGoPackageNotLongForm(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&fileOptionsGoPackageNotLongFormVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type fileOptionsGoPackageNotLongFormVisitor struct {
	baseAddVisitor

	option *proto.Option
}

func (v *fileOptionsGoPackageNotLongFormVisitor) OnStart(descriptor *FileDescriptor) error {
	v.option = nil
	return nil
}

func (v *fileOptionsGoPackageNotLongFormVisitor) VisitOption(element *proto.Option) {
	if element.Name == "go_package" {
		v.option = element
	}
}

func (v *fileOptionsGoPackageNotLongFormVisitor) Finally() error {
	if v.option == nil {
		return nil
	}
	value := v.option.Constant.Source
	if strings.Contains(value, ";") {
		v.AddFailuref(v.option.Position, `Option "go_package" cannot be of the long-form "go/import/path;package" and must only be of the short-form "package", but was %q.`, value)
	}
	return nil
}
