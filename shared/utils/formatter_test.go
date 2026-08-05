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
		{"Rounding Up", 1.006, 1.01},
		{"Rounding Down", 1.004, 1.00},
		{"Exact Integer", 100, 100},
		{"Negative Rounding Up", -1.006, -1.01},
		{"Negative Rounding Down", -1.004, -1.00},
		// .xx5 literals are not exact in float64; 1.005*100 is slightly under 100.5
		{"Positive half-ish", 1.005, 1.00},
		{"Negative half-ish", -1.005, -1.00},
		{"Large Number", 123456.78945345545, 123456.79},
		{"Small Number", 0.00001, 0},
		{"NaN", math.NaN(), math.NaN()},
		{"+Inf", math.Inf(1), math.Inf(1)},
		{"-Inf", math.Inf(-1), math.Inf(-1)},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := RoundToTwo(tt.input)
			if math.IsNaN(tt.expected) {
				s.True(math.IsNaN(result), "RoundToTwo(NaN) should be NaN")
				return
			}
			s.Equal(tt.expected, result, "RoundToTwo(%f) should equal %f", tt.input, tt.expected)
		})
	}
}

func (s *FormatterSuite) TestFormatTime() {
	tests := []struct {
		name     string
		input    float64
		unit     string
		round    bool
		expected float64
	}{
		{"Zero Input", 0, "ns", false, 0},
		{"Nanoseconds Default", 1000, "", false, 1000},
		{"Nanoseconds to Seconds", 5000000000, "s", false, 5},
		{"Nanoseconds to Milliseconds", 5000000, "ms", false, 5},
		{"Nanoseconds to Microseconds", 5000, "us", false, 5},
		{"Large Number", 5000000000000, "s", false, 5000},
		{"Small Number", 0.5, "ns", false, 0.5},
		{"Preserve sub-cent precision", 914273581, "s", false, 0.914273581},
		{"Round sub-cent precision", 914273581, "s", true, 0.91},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := FormatTime(tt.input, tt.unit, tt.round)
			s.Equal(tt.expected, result, "FormatTime(%f, %s, %v) should equal %f", tt.input, tt.unit, tt.round, tt.expected)
		})
	}
}

func (s *FormatterSuite) TestFormatMem() {
	tests := []struct {
		name     string
		input    float64
		unit     string
		round    bool
		expected float64
	}{
		{"Zero Input", 0, "B", false, 0},
		{"Bytes Default", 1024, "", false, 1024},
		{"Bytes to Bits", 64, "b", false, 512},
		{"Bytes to Kilobytes", 2048, "KB", false, 2},
		{"Bytes to Megabytes", 2097152, "MB", false, 2},
		{"Bytes to Gigabytes", 2147483648, "GB", false, 2},
		{"Large Number", 10737418240, "GB", false, 10},
		{"Small Number", 512, "KB", false, 0.5},
		{"Preserve fractional GB", 1, "GB", false, 1.0 / float64(byteToGb)},
		{"Round fractional GB", 1, "GB", true, 0},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := FormatMem(tt.input, tt.unit, tt.round)
			s.Equal(tt.expected, result, "FormatMem(%f, %s, %v) should equal %f", tt.input, tt.unit, tt.round, tt.expected)
		})
	}
}

func (s *FormatterSuite) TestFormatNumber() {
	tests := []struct {
		name     string
		input    float64
		unit     string
		round    bool
		expected float64
	}{
		{"Zero Input", 0, "", false, 0},
		{"Default Unit", 1000, "", false, 1000},
		{"Thousands (K)", 5000, "K", false, 5},
		{"Millions (M)", 5000000, "M", false, 5},
		{"Billions (B)", 5000000000, "B", false, 5},
		{"Trillions (T)", 5000000000000, "T", false, 5},
		{"Large Number", 10000000000, "B", false, 10},
		{"Small Number", 500, "K", false, 0.5},
		{"Full precision default", 0.914273581, "", false, 0.914273581},
		{"Round full precision", 0.914273581, "", true, 0.91},
		{"Fractional K no round", 999, "K", false, 0.999},
		{"Fractional K with round", 999, "K", true, 1.0},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := FormatNumber(tt.input, tt.unit, tt.round)
			s.Equal(tt.expected, result, "FormatNumber(%f, %s, %v) should equal %f", tt.input, tt.unit, tt.round, tt.expected)
		})
	}
}

func (s *FormatterSuite) TestFormatterEdgeCases() {
	s.Run("FormatTime Edge Cases", func() {
		s.Equal(veryLargeValue/nanoToSecond, FormatTime(veryLargeValue, "s", false), "Should handle very large values without rounding")
		s.Equal(RoundToTwo(veryLargeValue/nanoToSecond), FormatTime(veryLargeValue, "s", true), "Should round very large values when requested")

		s.Equal(-5.0, FormatTime(-5*nanoToSecond, "s", false), "Should handle negative values")
		s.Equal(smallValue, FormatTime(smallValue, "ns", false), "Should preserve very small values without rounding")
		s.Equal(0.0, FormatTime(smallValue, "ns", true), "Should round very small values to 0")

		s.Equal(1.0, FormatTime(nanoToSecond, "s", false), "Should convert exactly 1 second")
		s.Equal(1.0, FormatTime(nanoToMilli, "ms", false), "Should convert exactly 1 millisecond")
		s.Equal(1.0, FormatTime(nanoToMicro, "us", false), "Should convert exactly 1 microsecond")

		s.Equal(1000.0, FormatTime(1000, "invalid", false), "Should default to ns with invalid unit")
		s.Equal(0.0, FormatTime(0, "s", false), "Zero should remain zero for seconds")
		s.Equal(0.0, FormatTime(0, "ms", true), "Zero should remain zero for milliseconds")
	})

	s.Run("FormatMem Edge Cases", func() {
		s.Equal(1.0/float64(byteToGb), FormatMem(1, "GB", false), "Should preserve tiny fractional GB without rounding")
		s.Equal(0.0, FormatMem(1, "GB", true), "Should round tiny fractional GB to 0")

		s.Equal(veryLargeValue/byteToKb, FormatMem(veryLargeValue, "KB", false), "Should handle very large values without rounding")
		s.Equal(-2.0, FormatMem(-2*byteToKb, "KB", false), "Should handle negative values")
		s.Equal(smallValue, FormatMem(smallValue, "B", false), "Should preserve very small values without rounding")

		s.Equal(8.0, FormatMem(1, "b", false), "1 byte should equal 8 bits")
		s.Equal(1.0, FormatMem(byteToKb, "KB", false), "Should convert exactly 1 KB")
		s.Equal(1.0, FormatMem(byteToMb, "MB", false), "Should convert exactly 1 MB")
		s.Equal(1.0, FormatMem(byteToGb, "GB", false), "Should convert exactly 1 GB")

		s.Equal(1024.0, FormatMem(1024, "invalid", false), "Should default to bytes with invalid unit")
		s.Equal(0.0, FormatMem(0, "KB", false), "Zero should remain zero for kilobytes")
		s.Equal(0.0, FormatMem(0, "GB", true), "Zero should remain zero for gigabytes")
	})

	s.Run("FormatNumber Edge Cases", func() {
		s.Equal(veryLargeValue, FormatNumber(veryLargeValue, "", false), "Should handle very large values without rounding")
		s.Equal(RoundToTwo(veryLargeValue), FormatNumber(veryLargeValue, "", true), "Should round very large values when requested")

		s.Equal(-5.0, FormatNumber(-5*allocToK, "K", false), "Should handle negative values")
		s.Equal(smallValue, FormatNumber(smallValue, "", false), "Should preserve very small values without rounding")
		s.Equal(0.0, FormatNumber(smallValue, "", true), "Should round very small values to 0")

		s.Equal(1.0, FormatNumber(allocToK, "K", false), "Should convert exactly 1K allocations")
		s.Equal(1.0, FormatNumber(allocToM, "M", false), "Should convert exactly 1M allocations")
		s.Equal(1.0, FormatNumber(allocToB, "B", false), "Should convert exactly 1B allocations")
		s.Equal(1.0, FormatNumber(allocToT, "T", false), "Should convert exactly 1T allocations")

		s.Equal(1024.0, FormatNumber(1024, "invalid", false), "Should default to raw value with invalid unit")
		s.Equal(0.0, FormatNumber(0, "K", false), "Zero should remain zero for K")
		s.Equal(0.0, FormatNumber(0, "M", true), "Zero should remain zero for M")
	})
}

func (s *FormatterSuite) TestFormatterPrecision() {
	s.Run("FormatTime no round by default", func() {
		s.Equal(1.5, FormatTime(1.5*nanoToSecond, "s", false), "Should handle fractional values")
		s.Equal(0.001, FormatTime(nanoToMilli, "s", false), "Should preserve small fractional seconds without rounding")
		s.Equal(1.000001, FormatTime(1000001*nanoToMicro, "s", false), "Should preserve sub-cent precision")
		s.Equal(1.0, FormatTime(1000001*nanoToMicro, "s", true), "Should round to 2 decimal places when requested")
	})

	s.Run("FormatMem no round by default", func() {
		s.Equal(1.5, FormatMem(1.5*byteToKb, "KB", false), "Should handle fractional values")
		s.Equal(1024.0/float64(byteToMb), FormatMem(1024, "MB", false), "Should preserve small fractional MB without rounding")
		s.Equal(4096.0, FormatMem(512, "b", false), "Should precisely convert bytes to bits")
	})

	s.Run("FormatNumber no round by default", func() {
		s.Equal(1.5, FormatNumber(1.5*allocToK, "K", false), "Should handle fractional values")
		s.Equal(1e-6, FormatNumber(1, "M", false), "Should preserve very small fractional millions")
		s.Equal(1.000001, FormatNumber(1000001, "M", false), "Should preserve sub-cent precision")
		s.Equal(1.0, FormatNumber(1000001, "M", true), "Should round to 2 decimal places when requested")
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
					result := FormatTime(float64(j)*nanoToSecond, "s", false)
					results <- result
				}
			}()
		}

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
		round    bool
		expected float64
	}{
		{"Zero", 0, "ms", "ns", false, 0},
		{"Identity ns", 1000, "ns", "ns", false, 1000},
		{"Identity ms", 5.5, "ms", "ms", false, 5.5},
		{"ms to ns", 1, "ms", "ns", false, 1000000},
		{"ms to us", 1, "ms", "us", false, 1000},
		{"ms to s", 1000, "ms", "s", false, 1},
		{"ns to ms", 1000000, "ns", "ms", false, 1},
		{"ns to us", 1000, "ns", "us", false, 1},
		{"ns to s", 1e9, "ns", "s", false, 1},
		{"us to ms", 1000, "us", "ms", false, 1},
		{"us to ns", 1, "us", "ns", false, 1000},
		{"s to ms", 1, "s", "ms", false, 1000},
		{"s to us", 1, "s", "us", false, 1000000},
		{"s to ns", 1, "s", "ns", false, 1e9},
		{"Small ms to ns", 0.0038, "ms", "ns", false, 3800},
		{"Small ms to us", 0.0038, "ms", "us", false, 3.8},
		{"Identity with round", 1.006, "ms", "ms", true, 1.01},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := ConvertTime(tt.input, tt.from, tt.to, tt.round)
			s.Equal(tt.expected, result, "ConvertTime(%f, %s, %s, %v) should equal %f", tt.input, tt.from, tt.to, tt.round, tt.expected)
		})
	}
}

func (s *FormatterSuite) TestInputValidation() {
	s.Run("All Units Coverage", func() {
		timeUnits := []string{"", "ns", "us", "ms", "s"}
		for _, unit := range timeUnits {
			result := FormatTime(1000, unit, false)
			s.NotNil(result, "FormatTime should handle unit: %s", unit)
		}

		memUnits := []string{"", "B", "b", "KB", "MB", "GB"}
		for _, unit := range memUnits {
			result := FormatMem(1024, unit, false)
			s.NotNil(result, "FormatMem should handle unit: %s", unit)
		}

		numberUnits := []string{"", "K", "M", "B", "T"}
		for _, unit := range numberUnits {
			result := FormatNumber(1000, unit, false)
			s.NotNil(result, "FormatNumber should handle unit: %s", unit)
		}
	})

	s.Run("Case Sensitivity", func() {
		s.NotEqual(FormatNumber(1000, "k", false), FormatNumber(1000, "K", false), "Units should be case sensitive")
		s.NotEqual(FormatMem(1024, "kb", false), FormatMem(1024, "KB", false), "Units should be case sensitive")
	})
}

func TestFormatterSuite(t *testing.T) {
	suite.Run(t, new(FormatterSuite))
}
