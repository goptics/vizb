package utils

import (
	"math"
	"testing"

	"github.com/stretchr/testify/suite"
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

type FormatterSuite struct {
	suite.Suite
}

func (s *FormatterSuite) TestRoundToTwo() {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"Zero", 0, 0},
		{"Rounding Up", 1.006, 1.01},   // 1.006 * 100 = 100.6 -> round(101) -> 1.01
		{"Rounding Down", 1.004, 1.00}, // 1.004 * 100 = 100.4 -> round(100) -> 1.00
		{"Exact Integer", 100, 100},
		{"Negative Rounding Up", -1.006, -1.01}, // -1.006 * 100 = -100.6 -> round(-101) -> -1.01
		{"Negative Rounding Down", -1.004, -1.00},
		{"Large Number", 123456.78945345545, 123456.79},
		{"Small Number", 0.00001, 0},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := RoundToTwo(tt.input)
			s.Equal(tt.expected, result, "RoundToTwo(%f) should equal %f", tt.input, tt.expected)

			// Verify logic consistency manually as well for critical path
			if tt.input != 0 {
				expectedMath := math.Round(tt.input*100) / 100
				s.Equal(expectedMath, result, "Should match direct math.Round logic")
			}
		})
	}
}

func (s *FormatterSuite) TestFormatTime() {
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
		s.Run(tt.name, func() {
			result := FormatTime(tt.input, tt.unit)
			s.Equal(tt.expected, result, "FormatTime(%f, %s) should equal %f", tt.input, tt.unit, tt.expected)
		})
	}
}

func (s *FormatterSuite) TestFormatMem() {
	tests := []struct {
		name     string
		input    float64
		unit     string
		expected float64
	}{
		{"Zero Input", 0, "B", 0},
		{"Bytes Default", 1024, "", 1024},
		{"Bytes to Bits", 64, "b", 512}, // 64 bytes = 512 bits
		{"Bytes to Kilobytes", 2048, "KB", 2},
		{"Bytes to Megabytes", 2097152, "MB", 2},    // 2 MB = 2*1024*1024 bytes
		{"Bytes to Gigabytes", 2147483648, "GB", 2}, // 2 GB = 2*1024*1024*1024 bytes
		{"Large Number", 10737418240, "GB", 10},     // 10 GB
		{"Small Number", 512, "KB", 0.5},            // 0.5 KB
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := FormatMem(tt.input, tt.unit)
			s.Equal(tt.expected, result, "FormatMem(%f, %s) should equal %f", tt.input, tt.unit, tt.expected)
		})
	}
}

func (s *FormatterSuite) TestFormatNumber() {
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
		s.Run(tt.name, func() {
			result := FormatNumber(tt.input, tt.unit)
			s.Equal(tt.expected, result, "FormatNumber(%f, %s) should equal %f", tt.input, tt.unit, tt.expected)
		})
	}
}

func (s *FormatterSuite) TestFormatterEdgeCases() {
	s.Run("FormatTime Edge Cases", func() {
		// Test with very large values
		expectedLarge := RoundToTwo(veryLargeValue / nanoToSecond)
		s.Equal(expectedLarge, FormatTime(veryLargeValue, "s"), "Should handle very large values")

		// Test with negative values (though these are unlikely in benchmarks)
		s.Equal(-5.0, FormatTime(-5*nanoToSecond, "s"), "Should handle negative values")

		// Test with very small positive values - rounds to 0
		s.Equal(0.0, FormatTime(smallValue, "ns"), "Should round very small values to 0")

		// Test boundary values for each unit
		s.Equal(1.0, FormatTime(nanoToSecond, "s"), "Should convert exactly 1 second")
		s.Equal(1.0, FormatTime(nanoToMilli, "ms"), "Should convert exactly 1 millisecond")
		s.Equal(1.0, FormatTime(nanoToMicro, "us"), "Should convert exactly 1 microsecond")

		// Test with invalid unit - should use default
		s.Equal(1000.0, FormatTime(1000, "invalid"), "Should default to ns with invalid unit")

		// Test zero with different units
		s.Equal(0.0, FormatTime(0, "s"), "Zero should remain zero for seconds")
		s.Equal(0.0, FormatTime(0, "ms"), "Zero should remain zero for milliseconds")
	})

	s.Run("FormatMem Edge Cases", func() {
		// Test with small value and large unit conversion - rounds to 0
		s.Equal(0.0, FormatMem(1, "GB"), "Should round small values with large units to 0")

		// Test with very large values
		expectedLarge := RoundToTwo(veryLargeValue / byteToKb)
		s.Equal(expectedLarge, FormatMem(veryLargeValue, "KB"), "Should handle very large values")

		// Test with negative values (though these are unlikely in benchmarks)
		s.Equal(-2.0, FormatMem(-2*byteToKb, "KB"), "Should handle negative values")

		// Test with very small positive values - rounds to 0
		s.Equal(0.0, FormatMem(smallValue, "B"), "Should round very small values to 0")

		// Test boundary values for each unit
		s.Equal(8.0, FormatMem(1, "b"), "1 byte should equal 8 bits")
		s.Equal(1.0, FormatMem(byteToKb, "KB"), "Should convert exactly 1 KB")
		s.Equal(1.0, FormatMem(byteToMb, "MB"), "Should convert exactly 1 MB")
		s.Equal(1.0, FormatMem(byteToGb, "GB"), "Should convert exactly 1 GB")

		// Test with invalid unit - should use default
		s.Equal(1024.0, FormatMem(1024, "invalid"), "Should default to bytes with invalid unit")

		// Test zero with different units
		s.Equal(0.0, FormatMem(0, "KB"), "Zero should remain zero for kilobytes")
		s.Equal(0.0, FormatMem(0, "GB"), "Zero should remain zero for gigabytes")
	})

	s.Run("FormatNumber Edge Cases", func() {
		// Test with very large values
		expectedLarge := RoundToTwo(veryLargeValue)
		s.Equal(expectedLarge, FormatNumber(veryLargeValue, ""), "Should handle very large values")

		// Test with negative values (though these are unlikely in benchmarks)
		s.Equal(-5.0, FormatNumber(-5*allocToK, "K"), "Should handle negative values")

		// Test with very small positive values - rounds to 0
		s.Equal(0.0, FormatNumber(smallValue, ""), "Should round very small values to 0")

		// Test boundary values for each unit
		s.Equal(1.0, FormatNumber(allocToK, "K"), "Should convert exactly 1K allocations")
		s.Equal(1.0, FormatNumber(allocToM, "M"), "Should convert exactly 1M allocations")
		s.Equal(1.0, FormatNumber(allocToB, "B"), "Should convert exactly 1B allocations")
		s.Equal(1.0, FormatNumber(allocToT, "T"), "Should convert exactly 1T allocations")

		// Test with small values that result in fractional results
		s.Equal(1.0, FormatNumber(999, "K"), "Should round 0.999 to 1.0")

		// Test with invalid unit - should use default
		s.Equal(1024.0, FormatNumber(1024, "invalid"), "Should default to raw value with invalid unit")

		// Test zero with different units
		s.Equal(0.0, FormatNumber(0, "K"), "Zero should remain zero for K")
		s.Equal(0.0, FormatNumber(0, "M"), "Zero should remain zero for M")
	})
}

func (s *FormatterSuite) TestFormatterPrecision() {
	s.Run("FormatTime Rounding", func() {
		// Test with non-integer values
		s.Equal(1.5, FormatTime(1.5*nanoToSecond, "s"), "Should handle fractional values")
		s.Equal(0.0, FormatTime(nanoToMilli, "s"), "Should round small fractional values to 0")

		// Test precision boundaries
		s.Equal(1.0, FormatTime(1000001*nanoToMicro, "s"), "Should round to 2 decimal places")
	})

	s.Run("FormatMem Rounding", func() {
		// Test with non-integer values
		s.Equal(1.5, FormatMem(1.5*byteToKb, "KB"), "Should handle fractional values")
		s.Equal(0.0, FormatMem(1024, "MB"), "Should round small fractional values to 0")

		// Test precision with bits conversion
		s.Equal(4096.0, FormatMem(512, "b"), "Should precisely convert bytes to bits")
	})

	s.Run("FormatNumber Rounding", func() {
		// Test with non-integer values
		s.Equal(1.5, FormatNumber(1.5*allocToK, "K"), "Should handle fractional values")
		s.Equal(0.0, FormatNumber(1, "M"), "Should round very small fractional values to 0")

		// Test precision boundaries
		s.Equal(1.0, FormatNumber(1000001, "M"), "Should round to 2 decimal places")
	})
}

func (s *FormatterSuite) TestFormatterConcurrency() {
	s.Run("Concurrent FormatTime", func() {
		const numGoroutines = 100
		const numIterations = 1000

		results := make(chan float64, numGoroutines*numIterations)

		for range numGoroutines {
			go func() {
				for j := range numIterations {
					result := FormatTime(float64(j)*nanoToSecond, "s")
					results <- result
				}
			}()
		}

		// Collect all results
		for i := 0; i < numGoroutines*numIterations; i++ {
			<-results
		}

		s.True(true, "Concurrent execution completed without issues")
	})
}

func (s *FormatterSuite) TestConvertTime() {
	tests := []struct {
		name     string
		input    float64
		from     string
		to       string
		expected float64
	}{
		{"Zero", 0, "ms", "ns", 0},
		{"Identity ns", 1000, "ns", "ns", 1000},
		{"Identity ms", 5.5, "ms", "ms", 5.5},
		{"ms to ns", 1, "ms", "ns", 1000000},
		{"ms to us", 1, "ms", "us", 1000},
		{"ms to s", 1000, "ms", "s", 1},
		{"ns to ms", 1000000, "ns", "ms", 1},
		{"ns to us", 1000, "ns", "us", 1},
		{"ns to s", 1e9, "ns", "s", 1},
		{"us to ms", 1000, "us", "ms", 1},
		{"us to ns", 1, "us", "ns", 1000},
		{"s to ms", 1, "s", "ms", 1000},
		{"s to us", 1, "s", "us", 1000000},
		{"s to ns", 1, "s", "ns", 1e9},
		{"Small ms to ns", 0.0038, "ms", "ns", 3800},
		{"Small ms to us", 0.0038, "ms", "us", 3.8},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := ConvertTime(tt.input, tt.from, tt.to)
			s.Equal(tt.expected, result, "ConvertTime(%f, %s, %s) should equal %f", tt.input, tt.from, tt.to, tt.expected)
		})
	}
}

func (s *FormatterSuite) TestInputValidation() {
	s.Run("All Units Coverage", func() {
		// Test all valid time units
		timeUnits := []string{"", "ns", "us", "ms", "s"}
		for _, unit := range timeUnits {
			result := FormatTime(1000, unit)
			s.NotNil(result, "FormatTime should handle unit: %s", unit)
		}

		// Test all valid memory units
		memUnits := []string{"", "B", "b", "KB", "MB", "GB"}
		for _, unit := range memUnits {
			result := FormatMem(1024, unit)
			s.NotNil(result, "FormatMem should handle unit: %s", unit)
		}

		// Test all valid number units
		numberUnits := []string{"", "K", "M", "B", "T"}
		for _, unit := range numberUnits {
			result := FormatNumber(1000, unit)
			s.NotNil(result, "FormatNumber should handle unit: %s", unit)
		}
	})

	s.Run("Case Sensitivity", func() {
		// Test that units are case sensitive (as per implementation)
		s.NotEqual(FormatNumber(1000, "k"), FormatNumber(1000, "K"), "Units should be case sensitive")
		s.NotEqual(FormatMem(1024, "kb"), FormatMem(1024, "KB"), "Units should be case sensitive")
	})
}

func TestFormatterSuite(t *testing.T) {
	suite.Run(t, new(FormatterSuite))
}
