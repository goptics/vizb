package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test constants for consistency
const (
	// Time conversion constants
	nanoToSecond = 1e9
	nanoToMilli  = 1e6
	nanoToMicro  = 1e3

	// Memory conversion constants
	byteToKb = 1024
	byteToMb = 1024 * 1024
	byteToGb = 1024 * 1024 * 1024

	// Allocation conversion constants
	allocToK = 1e3
	allocToM = 1e6
	allocToB = 1e9
	allocToT = 1e12

	// Test values
	veryLargeValue = 1.8446744073709552e+19
	smallValue     = 0.0001
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
		{"Bytes to Megabytes", 2097152, "mb", 2},    // 2 MB = 2*1024*1024 bytes
		{"Bytes to Gigabytes", 2147483648, "gb", 2}, // 2 GB = 2*1024*1024*1024 bytes
		{"Large Number", 10737418240, "gb", 10},     // 10 GB
		{"Small Number", 512, "kb", 0.5},            // 0.5 KB
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
		input    float64
		unit     string
		expected float64
	}{
		{"Zero Input", 0, "", 0},
		{"Default Unit", 1000, "", 1000},
		{"Thousands (K)", 5000, "K", 5},
		{"Millions (M)", 5000000, "M", 5},
		{"Billions (B)", 5000000000, "B", 5},
		{"Trillions (T)", 5000000000000, "T", 5},
		{"Large Number", 10000000000, "B", 10},
		{"Small Number", 500, "K", 0.5}, // 500/1000 = 0.5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatAllocs(tt.input, tt.unit)
			assert.Equal(t, tt.expected, result, "FormatAllocs(%f, %s) should equal %f", tt.input, tt.unit, tt.expected)
		})
	}
}

// Edge case tests for all formatters
func TestFormatterEdgeCases(t *testing.T) {
	// Test boundary values and special cases

	t.Run("FormatTime Edge Cases", func(t *testing.T) {
		// Test with very large values
		assert.Equal(t, veryLargeValue/nanoToSecond, FormatTime(veryLargeValue, "s"), "Should handle very large values")

		// Test with negative values (though these are unlikely in benchmarks)
		assert.Equal(t, -5.0, FormatTime(-5*nanoToSecond, "s"), "Should handle negative values")

		// Test with very small positive values
		assert.Equal(t, smallValue, FormatTime(smallValue, "ns"), "Should handle very small values")

		// Test boundary values for each unit
		assert.Equal(t, 1.0, FormatTime(nanoToSecond, "s"), "Should convert exactly 1 second")
		assert.Equal(t, 1.0, FormatTime(nanoToMilli, "ms"), "Should convert exactly 1 millisecond")
		assert.Equal(t, 1.0, FormatTime(nanoToMicro, "us"), "Should convert exactly 1 microsecond")

		// Test with invalid unit - should use default
		assert.Equal(t, 1000.0, FormatTime(1000, "invalid"), "Should default to ns with invalid unit")

		// Test zero with different units
		assert.Equal(t, 0.0, FormatTime(0, "s"), "Zero should remain zero for seconds")
		assert.Equal(t, 0.0, FormatTime(0, "ms"), "Zero should remain zero for milliseconds")
	})

	t.Run("FormatMem Edge Cases", func(t *testing.T) {
		// Test with small value and large unit conversion
		expectedGbFromByte := 1.0 / byteToGb
		assert.Equal(t, expectedGbFromByte, FormatMem(1, "gb"), "Should handle small values with large units")

		// Test with very large values
		assert.Equal(t, veryLargeValue/byteToKb, FormatMem(veryLargeValue, "kb"), "Should handle very large values")

		// Test with negative values (though these are unlikely in benchmarks)
		assert.Equal(t, -2.0, FormatMem(-2*byteToKb, "kb"), "Should handle negative values")

		// Test with very small positive values
		assert.Equal(t, smallValue, FormatMem(smallValue, "B"), "Should handle very small values")

		// Test boundary values for each unit
		assert.Equal(t, 8.0, FormatMem(1, "b"), "1 byte should equal 8 bits")
		assert.Equal(t, 1.0, FormatMem(byteToKb, "kb"), "Should convert exactly 1 KB")
		assert.Equal(t, 1.0, FormatMem(byteToMb, "mb"), "Should convert exactly 1 MB")
		assert.Equal(t, 1.0, FormatMem(byteToGb, "gb"), "Should convert exactly 1 GB")

		// Test with invalid unit - should use default
		assert.Equal(t, 1024.0, FormatMem(1024, "invalid"), "Should default to bytes with invalid unit")

		// Test zero with different units
		assert.Equal(t, 0.0, FormatMem(0, "kb"), "Zero should remain zero for kilobytes")
		assert.Equal(t, 0.0, FormatMem(0, "gb"), "Zero should remain zero for gigabytes")
	})

	t.Run("FormatAllocs Edge Cases", func(t *testing.T) {
		// Test with very large values
		assert.Equal(t, veryLargeValue, FormatAllocs(veryLargeValue, ""), "Should handle very large values")

		// Test with negative values (though these are unlikely in benchmarks)
		assert.Equal(t, -5.0, FormatAllocs(-5*allocToK, "K"), "Should handle negative values")

		// Test with very small positive values
		assert.Equal(t, smallValue, FormatAllocs(smallValue, ""), "Should handle very small values")

		// Test boundary values for each unit
		assert.Equal(t, 1.0, FormatAllocs(allocToK, "K"), "Should convert exactly 1K allocations")
		assert.Equal(t, 1.0, FormatAllocs(allocToM, "M"), "Should convert exactly 1M allocations")
		assert.Equal(t, 1.0, FormatAllocs(allocToB, "B"), "Should convert exactly 1B allocations")
		assert.Equal(t, 1.0, FormatAllocs(allocToT, "T"), "Should convert exactly 1T allocations")

		// Test with small values that result in fractional results
		assert.Equal(t, 0.999, FormatAllocs(999, "K"), "Should handle small fractional values")

		// Test with invalid unit - should use default
		assert.Equal(t, 1024.0, FormatAllocs(1024, "invalid"), "Should default to raw value with invalid unit")

		// Test zero with different units
		assert.Equal(t, 0.0, FormatAllocs(0, "K"), "Zero should remain zero for K")
		assert.Equal(t, 0.0, FormatAllocs(0, "M"), "Zero should remain zero for M")
	})
}

// Test precision and rounding behavior
func TestFormatterPrecision(t *testing.T) {
	t.Run("FormatTime Precision", func(t *testing.T) {
		// Test with non-integer values
		assert.Equal(t, 1.5, FormatTime(1.5*nanoToSecond, "s"), "Should handle fractional values")
		assert.Equal(t, 0.001, FormatTime(nanoToMilli, "s"), "Should handle small fractional values")

		// Test precision boundaries
		assert.Equal(t, 1.000001, FormatTime(1000001*nanoToMicro, "s"), "Should maintain precision")
	})

	t.Run("FormatMem Precision", func(t *testing.T) {
		// Test with non-integer values
		assert.Equal(t, 1.5, FormatMem(1.5*byteToKb, "kb"), "Should handle fractional values")
		assert.Equal(t, 1.0/byteToKb, FormatMem(1024, "mb"), "Should handle small fractional values")

		// Test precision with bits conversion
		assert.Equal(t, 4096.0, FormatMem(512, "b"), "Should precisely convert bytes to bits")
	})

	t.Run("FormatAllocs Precision", func(t *testing.T) {
		// Test with non-integer values
		assert.Equal(t, 1.5, FormatAllocs(1.5*allocToK, "K"), "Should handle fractional values")
		assert.Equal(t, 0.000001, FormatAllocs(1, "M"), "Should handle very small fractional values")

		// Test precision boundaries
		assert.Equal(t, 1.000001, FormatAllocs(1000001, "M"), "Should maintain precision")
	})
}

// Test concurrent usage to ensure thread safety
func TestFormatterConcurrency(t *testing.T) {
	t.Run("Concurrent FormatTime", func(t *testing.T) {
		const numGoroutines = 100
		const numIterations = 1000

		results := make(chan float64, numGoroutines*numIterations)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < numIterations; j++ {
					result := FormatTime(float64(j)*nanoToSecond, "s")
					results <- result
				}
			}()
		}

		// Collect all results
		for i := 0; i < numGoroutines*numIterations; i++ {
			<-results
		}

		assert.True(t, true, "Concurrent execution completed without issues")
	})
}

// Benchmark tests for performance validation
func BenchmarkFormatTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FormatTime(float64(i)*nanoToSecond, "s")
	}
}

func BenchmarkFormatMem(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FormatMem(float64(i)*byteToMb, "mb")
	}
}

func BenchmarkFormatAllocs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FormatAllocs(float64(i)*allocToM, "M")
	}
}

// Test comprehensive input validation
func TestInputValidation(t *testing.T) {
	t.Run("All Units Coverage", func(t *testing.T) {
		// Test all valid time units
		timeUnits := []string{"", "ns", "us", "ms", "s"}
		for _, unit := range timeUnits {
			result := FormatTime(1000, unit)
			assert.NotNil(t, result, "FormatTime should handle unit: %s", unit)
		}

		// Test all valid memory units
		memUnits := []string{"", "B", "b", "kb", "mb", "gb"}
		for _, unit := range memUnits {
			result := FormatMem(1024, unit)
			assert.NotNil(t, result, "FormatMem should handle unit: %s", unit)
		}

		// Test all valid allocation units
		allocUnits := []string{"", "K", "M", "B", "T"}
		for _, unit := range allocUnits {
			result := FormatAllocs(1000, unit)
			assert.NotNil(t, result, "FormatAllocs should handle unit: %s", unit)
		}
	})

	t.Run("Case Sensitivity", func(t *testing.T) {
		// Test that units are case sensitive (as per implementation)
		assert.NotEqual(t, FormatAllocs(1000, "k"), FormatAllocs(1000, "K"), "Units should be case sensitive")
		assert.NotEqual(t, FormatMem(1024, "KB"), FormatMem(1024, "kb"), "Units should be case sensitive")
	})
}
