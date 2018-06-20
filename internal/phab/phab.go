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
	"fmt"

	"github.com/uber/prototool/internal/text"
)

const (
	// DefaultHarbormasterLintResultName is the default name
	// used when populating a HarbormasterLintResult.
	DefaultHarbormasterLintResultName = "PROTOTOOL"
	// DefaultHarbormasterLintResultCode is the default code
	// used when populating a HarbormasterLintResult. This will
	// only be used if there is no ID for lint failure.
	DefaultHarbormasterLintResultCode = "PROTOTOOL"
	// DefaultHarbormasterLintResultSeverity is the default severity
	// used when populating a HarbormasterLintResult.
	DefaultHarbormasterLintResultSeverity = "error"
)

// HarbormasterLintResult represents a text.Failure in a structure
// compatible with a Harbormaster Lint Result. It is meant to be
// encoded to JSON.
//
// https://secure.phabricator.com/conduit/method/harbormaster.sendmessage
type HarbormasterLintResult struct {
	Name        string `json:"name,omitempty"`
	Code        string `json:"code,omitempty"`
	Severity    string `json:"severity,omitempty"`
	Path        string `json:"path,omitempty"`
	Line        int    `json:"line,omitempty"`
	Char        int    `json:"char,omitempty"`
	Description string `json:"description,omitempty"`
}

// TextFailureToHarbormasterLintResult converts a text.Failure to a HarbormasterLintResult.
func TextFailureToHarbormasterLintResult(textFailure *text.Failure) (*HarbormasterLintResult, error) {
	if textFailure == nil {
		return nil, nil
	}
	if textFailure.Filename == "" {
		return nil, fmt.Errorf("%v could not be converted to a harbormaster lint result due to no Filename being set", textFailure)
	}
	harbormasterLintResult := &HarbormasterLintResult{
		Name:        DefaultHarbormasterLintResultName,
		Code:        textFailure.ID,
		Severity:    DefaultHarbormasterLintResultSeverity,
		Path:        textFailure.Filename,
		Line:        textFailure.Line,
		Char:        textFailure.Column,
		Description: textFailure.Message,
	}
	if harbormasterLintResult.Code == "" {
		harbormasterLintResult.Code = DefaultHarbormasterLintResultCode
	}
	return harbormasterLintResult, nil
}
