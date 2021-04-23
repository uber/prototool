// Copyright (c) 2021 Uber Technologies, Inc.
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

package text

import (
	"bytes"
	"testing"
	"text/scanner"

	"github.com/stretchr/testify/assert"
)

func TestFailureString(t *testing.T) {
	assert.Equal(t, "<input>:1:1:hello", newTestFailure("", 0, 0, "", "hello").String())
	assert.Equal(t, "<input>:1:2:hello", newTestFailure("", 0, 2, "", "hello").String())
	assert.Equal(t, "<input>:2:2:hello", newTestFailure("", 2, 2, "", "hello").String())
	assert.Equal(t, "foo:2:2:hello", newTestFailure("foo", 2, 2, "", "hello").String())
	assert.Equal(t, "foo:2:2:BAR hello", newTestFailure("foo", 2, 2, "BAR", "hello").String())
}

func TestFailureFprintln(t *testing.T) {
	testFailureFprintln(t, "2:1:<input>:BAR", newTestFailure("", 0, 2, "BAR", "hello"),
		FailureFieldColumn,
		FailureFieldLine,
		FailureFieldFilename,
		FailureFieldID,
	)
}

func testFailureFprintln(t *testing.T, expected string, failure *Failure, failureFields ...FailureField) {
	buffer := bytes.NewBuffer(nil)
	assert.NoError(t, failure.Fprintln(buffer, failureFields...))
	assert.Equal(t, expected+"\n", buffer.String())
}

func TestParseColonSeparatedFailureFields(t *testing.T) {
	testParseColonSeparatedFailureFields(t, "", false, DefaultFailureFields...)
	testParseColonSeparatedFailureFields(t, "filename", false, FailureFieldFilename)
	testParseColonSeparatedFailureFields(t, "filename:id", false, FailureFieldFilename, FailureFieldID)
	testParseColonSeparatedFailureFields(t, ":", true)
	testParseColonSeparatedFailureFields(t, ":filename:id", true)
	testParseColonSeparatedFailureFields(t, "filename:id:", true)
	testParseColonSeparatedFailureFields(t, "filename:id2", true)
}

func testParseColonSeparatedFailureFields(t *testing.T, input string, expectError bool, expectedFailureFields ...FailureField) {
	failureFields, err := ParseColonSeparatedFailureFields(input)
	if expectError {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
		assert.Equal(t, expectedFailureFields, failureFields)
	}
}

func TestSortFailures(t *testing.T) {
	failures := []*Failure{
		nil,
		newTestFailure("foo", 3, 3, "BAT", "mello"),
		newTestFailure("bar", 3, 3, "BAT", "mello"),
		newTestFailure("foo", 3, 3, "BAT", "hello"),
		newTestFailure("foo", 3, 3, "", "hello"),
		newTestFailure("foo", 2, 3, "", "hello"),
		newTestFailure("foo", 2, 2, "", "hello"),
		newTestFailure("foo", 2, 0, "", "hello"),
		newTestFailure("foo", 3, 3, "BAT", "mello"),
		newTestFailure("foo", 3, 3, "", "hello"),
		newTestFailure("foo", 0, 0, "", "hello"),
		newTestFailure("", 0, 0, "", "hello"),
		nil,
		nil,
		newTestFailure("foo", 3, 3, "BAT", "mello"),
		newTestFailure("foo", 3, 3, "BAT", "hello"),
		newTestFailure("foo", 3, 3, "BAR", "hello"),
		newTestFailure("foo", 2, 3, "", "hello"),
		newTestFailure("foo", 2, 4, "", "hello"),
		newTestFailure("foo", 2, 2, "", "hello"),
		newTestFailure("foo", 3, 3, "BAR", "hello"),
		newTestFailure("foo", 2, 0, "", "hello"),
		newTestFailure("foo", 0, 0, "", "hello"),
		newTestFailure("", 0, 0, "", "hello"),
		nil,
	}
	SortFailures(failures)
	assert.Equal(
		t,
		[]*Failure{
			nil,
			nil,
			nil,
			nil,
			newTestFailure("", 0, 0, "", "hello"),
			newTestFailure("", 0, 0, "", "hello"),
			newTestFailure("bar", 3, 3, "BAT", "mello"),
			newTestFailure("foo", 0, 0, "", "hello"),
			newTestFailure("foo", 0, 0, "", "hello"),
			newTestFailure("foo", 2, 0, "", "hello"),
			newTestFailure("foo", 2, 0, "", "hello"),
			newTestFailure("foo", 2, 2, "", "hello"),
			newTestFailure("foo", 2, 2, "", "hello"),
			newTestFailure("foo", 2, 3, "", "hello"),
			newTestFailure("foo", 2, 3, "", "hello"),
			newTestFailure("foo", 2, 4, "", "hello"),
			newTestFailure("foo", 3, 3, "", "hello"),
			newTestFailure("foo", 3, 3, "", "hello"),
			newTestFailure("foo", 3, 3, "BAR", "hello"),
			newTestFailure("foo", 3, 3, "BAR", "hello"),
			newTestFailure("foo", 3, 3, "BAT", "hello"),
			newTestFailure("foo", 3, 3, "BAT", "hello"),
			newTestFailure("foo", 3, 3, "BAT", "mello"),
			newTestFailure("foo", 3, 3, "BAT", "mello"),
			newTestFailure("foo", 3, 3, "BAT", "mello"),
		},
		failures,
	)
}

func newTestFailure(filename string, line int, column int, id string, message string) *Failure {
	return NewFailuref(scanner.Position{Filename: filename, Line: line, Column: column}, id, message)
}
