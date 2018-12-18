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
)

// Severity represents the signficance of a
// backward-incompatible change.
type Severity string

// The different levels of severity.
const (
	Source Severity = "source"
	Warn   Severity = "warn"
	Wire   Severity = "wire"
)

// Error represents an API-compatibility error.
type Error struct {
	// The full path to the filename, as in foo/bar.proto.
	Filename string   `json:"filename"`
	Line     int32    `json:"line"`
	Column   int32    `json:"column"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

// String returns a string representation of the Error type.
// The string is of the following form,
//
//  $(FILENAME):$(LINE):$(COLUMN):$(SEVERITY):$(MESSAGE)
//
//  For example,
//   "foo.proto:5:10:wire:Field number (1) with name "foo" was removed."
func (e Error) String() string {
	return fmt.Sprintf("%s:%d:%d:%s:%s", e.Filename, e.Line, e.Column, e.Severity, e.Message)
}

// Errors is defined in order to sort a slice of Errors.
type Errors []Error

// Less defines the precedence for sorting a slice of Errors.
// The order is as follows,
//  - Filename
//  - Line
//  - Column
//  - Severity
//  - Message
func (errs Errors) Less(i, j int) bool {
	if errs[i].Filename != errs[j].Filename {
		return errs[i].Filename < errs[j].Filename
	}
	if errs[i].Line != errs[j].Line {
		return errs[i].Line < errs[j].Line
	}
	if errs[i].Column != errs[j].Column {
		return errs[i].Column < errs[j].Column
	}
	if errs[i].Severity != errs[j].Severity {
		return errs[i].Severity < errs[j].Severity
	}
	return errs[i].Message < errs[j].Message
}

// Len returns the length of this slice of Errors.
func (errs Errors) Len() int {
	return len(errs)
}

// Swap swaps two elements in the slice of Errors.
func (errs Errors) Swap(i, j int) {
	errs[i], errs[j] = errs[j], errs[i]
}
