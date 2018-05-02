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

// Package phab provides functionality to interact with Phabricator.
//
// The primary purpose is to convert failures to JSON that is compatible
// with the Harbormaster API.
//
// https://secure.phabricator.com/conduit/method/harbormaster.sendmessage
package phab

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/prototool/internal/x/text"
)

func TestTextFailureToHarbormasterLintResult(t *testing.T) {
	assert.Equal(
		t,
		&HarbormasterLintResult{
			Name:        DefaultHarbormasterLintResultName,
			Code:        DefaultHarbormasterLintResultCode,
			Severity:    DefaultHarbormasterLintResultSeverity,
			Path:        "path/to/foo.proto",
			Line:        2,
			Description: "Foo is a foo.",
		},
		TextFailureToHarbormasterLintResult(
			&text.Failure{
				Filename: "path/to/foo.proto",
				Line:     2,
				Message:  "Foo is a foo.",
			},
		),
	)
	assert.Equal(
		t,
		&HarbormasterLintResult{
			Name:        DefaultHarbormasterLintResultName,
			Code:        "FOO",
			Severity:    DefaultHarbormasterLintResultSeverity,
			Path:        "path/to/foo.proto",
			Line:        2,
			Description: "Foo is a foo.",
		},
		TextFailureToHarbormasterLintResult(
			&text.Failure{
				Filename: "path/to/foo.proto",
				Line:     2,
				ID:       "FOO",
				Message:  "Foo is a foo.",
			},
		),
	)
}
