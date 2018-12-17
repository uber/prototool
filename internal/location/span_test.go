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
