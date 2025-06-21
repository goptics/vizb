package utils

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		unit     string
		expected float64
	}{
		{"Zero Input", 0, "ns", 0},
		{"Nanoseconds Default", 1000, "", 1000},
		{"Nanoseconds to Seconds", 5000000000, "s", 5},
		{"Nanoseconds to Milliseconds", 5000000, "ms", 5},
		{"Nanoseconds to Microseconds", 5000, "us", 5},
		{"Large Number", 5000000000000, "s", 5000},
		{"Small Number", 0.5, "ns", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTime(tt.input, tt.unit)
			assert.Equal(t, tt.expected, result, "FormatTime(%f, %s) should equal %f", tt.input, tt.unit, tt.expected)
		})
	}
}

func TestFormatMem(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		unit     string
		expected float64
	}{
		{"Zero Input", 0, "B", 0},
		{"Bytes Default", 1024, "", 1024},
		{"Bytes to Bits", 64, "b", 512}, // 64 bytes = 512 bits
		{"Bytes to Kilobytes", 2048, "kb", 2},
		{"Bytes to Megabytes", 2097152, "mb", 2}, // 2 MB = 2*1024*1024 bytes
		{"Bytes to Gigabytes", 2147483648, "gb", 2}, // 2 GB = 2*1024*1024*1024 bytes
		{"Large Number", 10737418240, "gb", 10}, // 10 GB
		{"Small Number", 512, "kb", 0.5}, // 0.5 KB
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatMem(tt.input, tt.unit)
			assert.Equal(t, tt.expected, result, "FormatMem(%f, %s) should equal %f", tt.input, tt.unit, tt.expected)
		})
	}
}

func TestFormatAllocs(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		unit     string
		expected uint64
	}{
		{"Zero Input", 0, "", 0},
		{"Default Unit", 1000, "", 1000},
		{"Thousands (K)", 5000, "K", 5},
		{"Millions (M)", 5000000, "M", 5},
		{"Billions (B)", 5000000000, "B", 5},
		{"Trillions (T)", 5000000000000, "T", 5},
		{"Large Number", 10000000000, "B", 10},
		{"Small Number", 500, "K", 0}, // 500/1000 = 0.5, truncated to 0 for uint64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatAllocs(tt.input, tt.unit)
			assert.Equal(t, tt.expected, result, "FormatAllocs(%d, %s) should equal %d", tt.input, tt.unit, tt.expected)
		})
	}
}

// Edge case tests for all formatters
func TestFormatterEdgeCases(t *testing.T) {
	// Test boundary values and special cases
	
	t.Run("FormatTime Edge Cases", func(t *testing.T) {
		// Test with very large values
		assert.Equal(t, 9.223372036854776e+9, FormatTime(9.223372036854776e+18, "s"), "Should handle very large values")
		
		// Test with negative values (though these are unlikely in benchmarks)
		assert.Equal(t, -5.0, FormatTime(-5000000000, "s"), "Should handle negative values")
		
		// Test with invalid unit - should use default
		assert.Equal(t, 1000.0, FormatTime(1000, "invalid"), "Should default to ns with invalid unit")
	})
	
	t.Run("FormatMem Edge Cases", func(t *testing.T) {
		// Test with small value and large unit conversion
		// 1 byte = 9.313225746154785e-10 GB (1/(1024^3))
		assert.Equal(t, 9.313225746154785e-10, FormatMem(1, "gb"), "Should handle small values with large units")
		
		// Test with negative values (though these are unlikely in benchmarks)
		assert.Equal(t, -2.0, FormatMem(-2048, "kb"), "Should handle negative values")
		
		// Test with invalid unit - should use default
		assert.Equal(t, 1024.0, FormatMem(1024, "invalid"), "Should default to bytes with invalid unit")
	})
	
	t.Run("FormatAllocs Edge Cases", func(t *testing.T) {
		// Test with maximum uint64 value
		maxUint64 := uint64(18446744073709551615)
		assert.Equal(t, maxUint64, FormatAllocs(maxUint64, ""), "Should handle max uint64 value")
		
		// Test with small values that get truncated to zero
		assert.Equal(t, uint64(0), FormatAllocs(999, "K"), "Should truncate values less than unit")
		
		// Test with invalid unit - should use default
		assert.Equal(t, uint64(1024), FormatAllocs(1024, "invalid"), "Should default to raw value with invalid unit")
	})
}

// Test precision and rounding behavior
func TestFormatterPrecision(t *testing.T) {
	t.Run("FormatTime Precision", func(t *testing.T) {
		// Test with non-integer values
		assert.Equal(t, 1.5, FormatTime(1500000000, "s"), "Should handle fractional values")
		assert.Equal(t, 0.001, FormatTime(1000000, "s"), "Should handle small fractional values")
	})
	
	t.Run("FormatMem Precision", func(t *testing.T) {
		// Test with non-integer values
		assert.Equal(t, 1.5, FormatMem(1536, "kb"), "Should handle fractional values")
		// 1024 bytes = 0.0009765625 MB (1024/(1024^2))
		assert.Equal(t, 0.0009765625, FormatMem(1024, "mb"), "Should handle small fractional values")
	})
}
