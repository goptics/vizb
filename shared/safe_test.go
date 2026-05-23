package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithSafe(t *testing.T) {
	t.Run("No Panic", func(t *testing.T) {
		executed := false
		err := WithSafe("test function", func() {
			executed = true
		})

		assert.True(t, executed, "Expected function to be executed")
		assert.NoError(t, err)
	})

	t.Run("Panic Recovery", func(t *testing.T) {
		err := WithSafe("test panic", func() {
			panic("test panic message")
		})

		require.Error(t, err, "Expected error from panic recovery")

		expectedMsg := "panic recovered inside test panic: test panic message"
		assert.EqualError(t, err, expectedMsg)
	})

	t.Run("Panic with Different Types", func(t *testing.T) {
		err := WithSafe("string panic", func() {
			panic("string error")
		})
		assert.Error(t, err, "Expected error from string panic")

		err = WithSafe("int panic", func() {
			panic(42)
		})
		assert.Error(t, err, "Expected error from int panic")
	})
}
