package location

import (
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
)

func TestFinder(t *testing.T) {
	tests := []struct {
		desc         string
		givePath     []int32
		giveSpan     []int32
		giveDetached []string
		giveLeading  string
		giveTrailing string
		giveID       ID
		wantLine     int32
		wantCol      int32
		found        bool
	}{
		{
			desc:     "Path not found",
			givePath: []int32{1},
			found:    false,
		},
		{
			desc:         "Top-level message",
			givePath:     []int32{4, 0},
			giveSpan:     []int32{8, 0},
			giveDetached: []string{"Leading, detached comment"},
			giveLeading:  "Leading comment",
			giveTrailing: "Trailing comment",
			giveID:       Message,
			wantLine:     9,
			wantCol:      1,
			found:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var basePath Path
			src := &descriptor.SourceCodeInfo{
				Location: []*descriptor.SourceCodeInfo_Location{
					{
						Path:                    tt.givePath,
						Span:                    tt.giveSpan,
						LeadingComments:         &tt.giveLeading,
						TrailingComments:        &tt.giveTrailing,
						LeadingDetachedComments: tt.giveDetached,
					},
				},
			}
			f := NewFinder(src)
			l, ok := f.Find(basePath.Scope(tt.giveID, 0))

			if !tt.found {
				assert.False(t, tt.found)
				return
			}
			assert.True(t, ok)
			assert.Equal(t, tt.wantLine, l.Span.Line())
			assert.Equal(t, tt.wantCol, l.Span.Col())
			assert.Equal(t, tt.giveDetached, l.Comments.LeadingDetached)
			assert.Equal(t, tt.giveLeading, l.Comments.Leading)
			assert.Equal(t, tt.giveTrailing, l.Comments.Trailing)
		})
	}
}
