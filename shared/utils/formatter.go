package utils

import "math"

// float64Epsilon is JS Number.EPSILON (2^-52), used so RoundToTwo matches the
// former UI round2 half-away-from-zero tie correction on binary floats.
const float64Epsilon = 2.220446049250313e-16

// RoundToTwo rounds to 2 decimal places with the same decimal-aware epsilon
// nudge as the previous UI round2: exact ties (e.g. 1.005 → 1.01, -1.005 → -1.01)
// round half away from zero instead of drifting on binary float error.
// Non-finite values are returned unchanged.
func RoundToTwo(num float64) float64 {
	if math.IsNaN(num) || math.IsInf(num, 0) {
		return num
	}

	scaled := num * 100
	sign := 1.0
	switch {
	case scaled < 0:
		sign = -1
	case scaled == 0:
		sign = 0
	}
	mag := math.Abs(scaled)
	if mag < 1 {
		mag = 1
	}
	tieCorrected := scaled + sign*float64Epsilon*mag
	return math.Round(tieCorrected) / 100
}

// FormatTime converts a time value from nanoseconds to the specified unit.
// Supported units: "ns" (nanoseconds), "us" (microseconds), "ms" (milliseconds), "s" (seconds).
// When round is true, the converted value is rounded to 2 decimals via RoundToTwo.
func FormatTime(n float64, unit string, round bool) (time float64) {
	if n == 0 {
		return
	}

	switch unit {
	case "s":
		time = n / 1e9
	case "ms":
		time = n / 1e6
	case "us":
		time = n / 1e3
	default:
		time = n
	}

	if round {
		return RoundToTwo(time)
	}
	return time
}

// FormatMem converts a memory value from bytes to the specified unit.
// Supported units: "b" (bits), "B" (bytes), "KB" (kilobytes), "MB" (megabytes), "GB" (gigabytes).
// For "b" unit, bytes are converted to bits by multiplying by 8.
// When round is true, the converted value is rounded to 2 decimals via RoundToTwo.
func FormatMem(n float64, unit string, round bool) (mem float64) {
	if n == 0 {
		return
	}

	switch unit {
	case "b":
		mem = n * 8
	case "KB":
		mem = n / 1024
	case "MB":
		mem = n / (1024 * 1024)
	case "GB":
		mem = n / (1024 * 1024 * 1024)
	default:
		mem = n
	}

	if round {
		return RoundToTwo(mem)
	}
	return mem
}

// ConvertTime converts a time value from one unit to another.
// Supported units: "ns", "us", "ms", "s".
// When round is true, the converted value is rounded to 2 decimals via RoundToTwo.
func ConvertTime(n float64, from, to string, round bool) float64 {
	if n == 0 {
		return 0
	}

	if from == to {
		if round {
			return RoundToTwo(n)
		}
		return n
	}

	var ns float64
	switch from {
	case "s":
		ns = n * 1e9
	case "ms":
		ns = n * 1e6
	case "us":
		ns = n * 1e3
	default:
		ns = n
	}

	return FormatTime(ns, to, round)
}

// FormatNumber converts an allocation count (or generic metric) to the specified unit.
// Supported units: "K" (thousands), "M" (millions), "B" (billions), "T" (trillions).
// Empty unit string returns the value unchanged (aside from optional rounding).
// When round is true, the result is rounded to 2 decimals via RoundToTwo.
func FormatNumber(n float64, unit string, round bool) (allocs float64) {
	if n == 0 {
		return
	}

	switch unit {
	case "K":
		allocs = n / 1e3
	case "M":
		allocs = n / 1e6
	case "B":
		allocs = n / 1e9
	case "T":
		allocs = n / 1e12
	default:
		allocs = n
	}

	if round {
		return RoundToTwo(allocs)
	}
	return allocs
}
