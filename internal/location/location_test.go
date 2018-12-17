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
