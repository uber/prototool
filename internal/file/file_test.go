package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAbsClean(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	tests := []struct {
		desc string
		give string
		want string
	}{
		{
			desc: "Empty path",
		},
		{
			desc: "Root path",
			give: "/",
			want: "/",
		},
		{
			desc: "Clean absolute path",
			give: "/path/to/foo",
			want: "/path/to/foo",
		},
		{
			desc: "Messy absolute path",
			give: "/path//to///foo",
			want: "/path/to/foo",
		},
		{
			desc: "Clean relative path",
			give: "path/to/foo",
			want: filepath.Join(wd, "path/to/foo"),
		},
		{
			desc: "Messy relative path",
			give: "path//to///foo",
			want: filepath.Join(wd, "path/to/foo"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			res, err := AbsClean(tt.give)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestCheckAbs(t *testing.T) {
	t.Run("Absolute path", func(t *testing.T) {
		assert.NoError(t, CheckAbs("/path/to/foo"))
	})
	t.Run("Relative path", func(t *testing.T) {
		assert.EqualError(t, CheckAbs("path/to/foo"), "expected absolute path but was path/to/foo")
	})
}
