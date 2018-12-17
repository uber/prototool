package compatible

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// checkerFn is is a generic representation for compatibility
// check functions.
type checkerFn func(descriptorProtoGroup, descriptorProtoGroup)

// check executes the given fn, and asserts that the expected error,
// if any, is equivalent to the one found on the fileChecker.
func check(t *testing.T, c *fileChecker, fn checkerFn, original, updated descriptorProtoGroup, err string) {
	fn(original, updated)
	if err == "" {
		assert.Empty(t, c.errors)
		return
	}
	require.Len(t, c.errors, 1)
	assert.Equal(t, c.errors[0].String(), err)
}
