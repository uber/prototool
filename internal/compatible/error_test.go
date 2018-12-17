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
