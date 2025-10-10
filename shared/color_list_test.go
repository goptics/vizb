package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNextColorFor(t *testing.T) {
	t.Run("Returns same color for same key", func(t *testing.T) {
		resetColorState()

		color1 := GetNextColorFor("key1")
		color2 := GetNextColorFor("key1")
		color3 := GetNextColorFor("key1")

		assert.Equal(t, color1, color2)
		assert.Equal(t, color2, color3)
		assert.Contains(t, ColorList, color1)
	})

	t.Run("Returns different colors for different keys", func(t *testing.T) {
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

	t.Run("Handles empty string key consistently", func(t *testing.T) {
		resetColorState()

		color1 := GetNextColorFor("")
		color2 := GetNextColorFor("")

		assert.NotEmpty(t, color1)
		assert.Contains(t, ColorList, color1)
		assert.Equal(t, color1, color2, "Empty key should return same color")
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
		firstColors := make([]string, colorListLen)

		// Get first full cycle of colors
		for i := range firstColors {
			firstColors[i] = GetNextColorFor(stringFromInt(i))
		}

		// Get colors beyond the list length (should wrap)
		extraColors := make([]string, 5)
		for i := range extraColors {
			extraColors[i] = GetNextColorFor(stringFromInt(colorListLen + i))
		}

		// All returned colors should be valid
		for i, color := range firstColors {
			assert.Contains(t, ColorList, color, "Color at index %d should be valid", i)
		}
		for i, color := range extraColors {
			assert.Contains(t, ColorList, color, "Wrapped color at index %d should be valid", i)
		}

		// Verify that requesting same keys returns same colors
		assert.Equal(t, firstColors[0], GetNextColorFor(stringFromInt(0)))
		assert.Equal(t, extraColors[0], GetNextColorFor(stringFromInt(colorListLen)))
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
