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
	"text/scanner"

	"github.com/uber/prototool/internal/text"
)

var fileHeaderLinter = NewLinter(
	"FILE_HEADER",
	"Verifies that the file header matches the expected file header if the file_header option is set in the configuration file.",
	checkFileHeader,
)

func checkFileHeader(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(fileHeaderVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type fileHeaderVisitor struct {
	baseAddVisitor
}

func (v fileHeaderVisitor) OnStart(descriptor *FileDescriptor) error {
	if fileHeader := descriptor.ProtoSet.Config.Lint.FileHeader; fileHeader != "" {
		if !strings.HasPrefix(descriptor.FileData, fileHeader) {
			v.AddFailuref(scanner.Position{Filename: descriptor.Filename}, "File is expected to begin with the file header specified from the configuration file.")
		}
	}
	return nil
}
