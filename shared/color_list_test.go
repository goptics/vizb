package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNextColorFor(t *testing.T) {
	t.Run("Returns colors sequentially for different keys", func(t *testing.T) {
		resetColorState()

		color1 := GetNextColorFor("key1")
		color2 := GetNextColorFor("key2")
		color3 := GetNextColorFor("key3")

		assert.NotEqual(t, color1, color2)
		assert.NotEqual(t, color2, color3)
		assert.Contains(t, ColorList, color1)
		assert.Contains(t, ColorList, color2)
		assert.Contains(t, ColorList, color3)
	})

	t.Run("Handles empty string key", func(t *testing.T) {
		resetColorState()

		color := GetNextColorFor("")
		assert.NotEmpty(t, color)
		assert.Contains(t, ColorList, color)
	})

	t.Run("Handles special characters in keys", func(t *testing.T) {
		resetColorState()

		keys := []string{
			"key with spaces",
			"key!@#$%",
			"key\nwith\nnewlines",
		}

		for _, key := range keys {
			color := GetNextColorFor(key)
			assert.NotEmpty(t, color)
			assert.Contains(t, ColorList, color)
		}
	})

	t.Run("Wraps around after exhausting colors", func(t *testing.T) {
		resetColorState()

		colorListLen := len(ColorList)
		colors := make([]string, colorListLen+5)

		for i := range colors {
			colors[i] = GetNextColorFor(stringFromInt(i))
		}

		// All returned colors should be valid
		for i, color := range colors {
			assert.Contains(t, ColorList, color, "Color at index %d should be valid", i)
		}
	})

	t.Run("Different similar keys get different colors", func(t *testing.T) {
		resetColorState()

		color1 := GetNextColorFor("test")
		color2 := GetNextColorFor("test1")
		color3 := GetNextColorFor("Test")

		assert.NotEqual(t, color1, color2)
		assert.NotEqual(t, color1, color3)
		assert.NotEqual(t, color2, color3)
	})
}

func TestColorListIntegrity(t *testing.T) {
	t.Run("All colors are valid hex format", func(t *testing.T) {
		for i, color := range ColorList {
			assert.Regexp(t, "^#[0-9A-Fa-f]{6}$", color, "Color at index %d should be valid hex", i)
		}
	})

	t.Run("No duplicate colors", func(t *testing.T) {
		colorSet := make(map[string]bool)

		for _, color := range ColorList {
			assert.False(t, colorSet[color], "Color %s appears multiple times", color)
			colorSet[color] = true
		}
	})

	t.Run("ColorList is not empty", func(t *testing.T) {
		assert.NotEmpty(t, ColorList)
		assert.Greater(t, len(ColorList), 0)
	})
}

// Helper functions

func resetColorState() {
	colorMap = make(map[string]int)
	i = 0
}

func stringFromInt(n int) string {
	return string(rune('a'+(n%26))) + string(rune('0'+(n/26)))
}
