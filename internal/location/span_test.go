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

	"github.com/stretchr/testify/assert"
)

func TestSpan(t *testing.T) {
	tests := []struct {
		desc       string
		giveSpan   []int32
		wantLine   int32
		wantCol    int32
		wantString string
	}{
		{
			desc:       "Default value",
			wantLine:   1,
			wantCol:    1,
			wantString: "1:1",
		},
		{
			desc:       "Equal line, col",
			giveSpan:   []int32{1, 1},
			wantLine:   2,
			wantCol:    2,
			wantString: "2:2",
		},
		{
			desc:       "Unequal line, col",
			giveSpan:   []int32{1, 2},
			wantLine:   2,
			wantCol:    3,
			wantString: "2:3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := NewSpan(tt.giveSpan)
			assert.Equal(t, tt.wantLine, s.Line())
			assert.Equal(t, tt.wantCol, s.Col())
			assert.Equal(t, tt.wantString, s.String())
		})
	}
}
