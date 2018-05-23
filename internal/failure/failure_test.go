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

package failure

import (
	"bytes"
	"testing"
	"text/scanner"

	"github.com/stretchr/testify/assert"
)

func TestIDString(t *testing.T) {
	assert.Equal(t, "ERROR", Error.String())
	assert.Equal(t, "FORMAT", Format.String())
	assert.Equal(t, "LINT", Lint.String())
	assert.Equal(t, "PROTO", Proto.String())
}

func TestFailureFprintln(t *testing.T) {
	testFailureFprintln(t, "2:1:<input>:LINT", newTestFailure("", 0, 2, Lint, "hello"),
		Column,
		Line,
		Filename,
		Identifier,
	)
}

func testFailureFprintln(t *testing.T, expected string, failure *Failure, failureFields ...Field) {
	buffer := bytes.NewBuffer(nil)
	assert.NoError(t, failure.Fprintln(buffer, failureFields...))
	assert.Equal(t, expected+"\n", buffer.String())
}

func TestParseFields(t *testing.T) {
	testParseFields(t, "", false, _defaultFields...)
	testParseFields(t, "filename", false, Filename)
	testParseFields(t, "filename:id", false, Filename, Identifier)
	testParseFields(t, ":", true)
	testParseFields(t, ":filename:id", true)
	testParseFields(t, "filename:id:", true)
	testParseFields(t, "filename:id2", true)
}

func testParseFields(t *testing.T, input string, expectError bool, expectedFields ...Field) {
	failureFields, err := ParseFields(input)
	if expectError {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
		assert.Equal(t, expectedFields, failureFields)
	}
}

func TestSort(t *testing.T) {
	failures := []*Failure{
		nil,
		newTestFailure("foo", 3, 3, Proto, "mello"),
		newTestFailure("bar", 3, 3, Proto, "mello"),
		newTestFailure("foo", 3, 3, Proto, "hello"),
		newTestFailure("foo", 3, 3, Format, "hello"),
		newTestFailure("foo", 2, 3, Format, "hello"),
		newTestFailure("foo", 2, 2, Format, "hello"),
		newTestFailure("foo", 2, 0, Format, "hello"),
		newTestFailure("foo", 3, 3, Proto, "mello"),
		newTestFailure("foo", 3, 3, Format, "hello"),
		newTestFailure("foo", 0, 0, Format, "hello"),
		newTestFailure("", 0, 0, Format, "hello"),
		nil,
		nil,
		newTestFailure("foo", 3, 3, Proto, "mello"),
		newTestFailure("foo", 3, 3, Proto, "hello"),
		newTestFailure("foo", 3, 3, Lint, "hello"),
		newTestFailure("foo", 2, 3, Format, "hello"),
		newTestFailure("foo", 2, 4, Format, "hello"),
		newTestFailure("foo", 2, 2, Format, "hello"),
		newTestFailure("foo", 3, 3, Lint, "hello"),
		newTestFailure("foo", 2, 0, Format, "hello"),
		newTestFailure("foo", 0, 0, Format, "hello"),
		newTestFailure("", 0, 0, Format, "hello"),
		nil,
	}
	Sort(failures)
	assert.Equal(
		t,
		[]*Failure{
			nil,
			nil,
			nil,
			nil,
			newTestFailure("", 0, 0, Format, "hello"),
			newTestFailure("", 0, 0, Format, "hello"),
			newTestFailure("bar", 3, 3, Proto, "mello"),
			newTestFailure("foo", 0, 0, Format, "hello"),
			newTestFailure("foo", 0, 0, Format, "hello"),
			newTestFailure("foo", 2, 0, Format, "hello"),
			newTestFailure("foo", 2, 0, Format, "hello"),
			newTestFailure("foo", 2, 2, Format, "hello"),
			newTestFailure("foo", 2, 2, Format, "hello"),
			newTestFailure("foo", 2, 3, Format, "hello"),
			newTestFailure("foo", 2, 3, Format, "hello"),
			newTestFailure("foo", 2, 4, Format, "hello"),
			newTestFailure("foo", 3, 3, Format, "hello"),
			newTestFailure("foo", 3, 3, Format, "hello"),
			newTestFailure("foo", 3, 3, Lint, "hello"),
			newTestFailure("foo", 3, 3, Lint, "hello"),
			newTestFailure("foo", 3, 3, Proto, "hello"),
			newTestFailure("foo", 3, 3, Proto, "hello"),
			newTestFailure("foo", 3, 3, Proto, "mello"),
			newTestFailure("foo", 3, 3, Proto, "mello"),
			newTestFailure("foo", 3, 3, Proto, "mello"),
		},
		failures,
	)
}

func newTestFailure(filename string, line int, column int, id ID, message string) *Failure {
	return Newf(scanner.Position{Filename: filename, Line: line, Column: column}, id, message)
}
