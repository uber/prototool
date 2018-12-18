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

package location

import (
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
)

func newTestSourceCodeLocation(
	path []int32,
	span []int32,
	detached []string,
	leading string,
	trailing string,
) *descriptor.SourceCodeInfo_Location {
	return &descriptor.SourceCodeInfo_Location{
		Path:                    path,
		Span:                    span,
		LeadingDetachedComments: detached,
		LeadingComments:         &leading,
		TrailingComments:        &trailing,
	}
}

func TestLocation(t *testing.T) {
	tests := []struct {
		desc         string
		giveSpan     []int32
		giveDetached []string
		giveLeading  string
		giveTrailing string
		wantLine     int32
		wantCol      int32
	}{
		{
			desc:     "Default value location",
			wantLine: 1,
			wantCol:  1,
		},
		{
			desc:         "Hydrated location",
			giveSpan:     []int32{1, 2},
			giveDetached: []string{"Leading, detached comment"},
			giveLeading:  "Leading comment",
			giveTrailing: "Trailing comment",
			wantLine:     2,
			wantCol:      3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			l := NewLocation(
				newTestSourceCodeLocation(
					nil, // Path
					tt.giveSpan,
					tt.giveDetached,
					tt.giveLeading,
					tt.giveTrailing,
				),
			)
			assert.Equal(t, tt.wantLine, l.Span.Line())
			assert.Equal(t, tt.wantCol, l.Span.Col())
			assert.Equal(t, tt.giveDetached, l.Comments.LeadingDetached)
			assert.Equal(t, tt.giveLeading, l.Comments.Leading)
			assert.Equal(t, tt.giveTrailing, l.Comments.Trailing)
		})
	}
}
