package location

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	var base Path
	t.Run("Scope", func(t *testing.T) {
		foo := base.Scope(Message, 0)
		bar := base.Scope(Message, 0)
		assert.Equal(t, foo, bar)

		// Updating one should not change the other.
		foo = foo.Scope(Field, 0)
		assert.Equal(t, Path{4, 0, 2, 0}, foo)
		assert.NotEqual(t, foo, bar)
	})
	t.Run("Target", func(t *testing.T) {
		foo := base.Target(Name)
		bar := base.Target(Name)
		assert.Equal(t, foo, bar)

		// Updating one should not change the other.
		foo = foo.Target(Name)
		assert.Equal(t, Path{1, 1}, foo)
		assert.NotEqual(t, foo, bar)
	})
}
