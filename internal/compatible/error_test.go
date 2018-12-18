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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	tests := []struct {
		desc     string
		filename string
		line     int32
		column   int32
		severity Severity
		message  string
		wantErr  string
	}{
		{
			desc:     "Error with default Span",
			filename: _testFilename,
			line:     1,
			column:   1,
			severity: Wire,
			message:  "Unable to locate error!",
			wantErr:  "test.proto:1:1:wire:Unable to locate error!",
		},
		{
			desc:     "Error with non-zero Span",
			filename: _testFilename,
			line:     6,
			column:   11,
			severity: Source,
			message:  "Error located!",
			wantErr:  "test.proto:6:11:source:Error located!",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			err := Error{
				Filename: tt.filename,
				Line:     tt.line,
				Column:   tt.column,
				Severity: tt.severity,
				Message:  tt.message,
			}
			assert.Equal(t, err.String(), tt.wantErr)
		})
	}
}
