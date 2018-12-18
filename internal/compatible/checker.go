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

package compatible

import (
	"fmt"
	"sort"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/location"
)

// fileChecker collects compatibility errors for a specific file.
type fileChecker struct {
	filename string
	finder   *location.Finder
	errors   Errors
}

func newFileChecker(from *descriptor.FileDescriptorProto) *fileChecker {
	return &fileChecker{
		filename: from.GetName(),
		finder:   location.NewFinder(from.GetSourceCodeInfo()),
	}
}

// Check determines whether "to" is backward-compatible with "from".
func Check(from, to *descriptor.FileDescriptorSet) []Error {
	var (
		basePath location.Path
		errs     Errors
	)

	fs := newFileSet(from)
	for _, updated := range to.GetFile() {
		if original, ok := fs[updated.GetName()]; ok {
			c := newFileChecker(original.descriptor)
			errs = append(errs, c.checkFile(original, newFile(updated, basePath))...)
		}
	}

	fs = newFileSet(to)
	for _, original := range from.GetFile() {
		filename := original.GetName()
		if _, ok := fs[filename]; !ok {
			errs = append(
				errs,
				Error{
					Filename: filename,
					Severity: Warn,
					Message:  fmt.Sprintf("Failed to validate file %q: the file no longer exists.", filename),
				},
			)
		}
	}

	sort.Sort(errs)
	return errs
}

// AddErrorf adds an Error to the fileChecker using the given path and formatted message.
func (c *fileChecker) AddErrorf(path location.Path, severity Severity, format string, args ...interface{}) {
	// If a location does not exist for
	// the given path, we fall back to
	// the default location.Span value.
	loc, _ := c.finder.Find(path)

	err := Error{
		Filename: c.filename,
		Line:     loc.Span.Line(),
		Column:   loc.Span.Col(),
		Severity: severity,
		Message:  fmt.Sprintf(format, args...),
	}
	c.errors = append(c.errors, err)
}
