package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExternalConfigValidate(t *testing.T) {
	t.Run("no_default_excludes", func(t *testing.T) {
		e := ExternalConfig{NoDefaultExcludes: true}
		assert.EqualError(t, e.Validate(), "no_default_excludes is not a configurable setting")
	})
}
