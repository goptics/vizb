package utils

// FormatTime converts a time value from nanoseconds to the specified unit.
// Supported units: "ns" (nanoseconds), "us" (microseconds), "ms" (milliseconds), "s" (seconds).
// Returns the converted time value as a float64.
func FormatTime(n float64, unit string) (time float64) {
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

	return
}

// FormatMem converts a memory value from bytes to the specified unit.
// Supported units: "b" (bits), "B" (bytes), "KB" (kilobytes), "MB" (megabytes), "GB" (gigabytes).
// For "b" unit, bytes are converted to bits by multiplying by 8.
// Returns the converted memory value as a float64.
func FormatMem(n float64, unit string) (mem float64) {
	if n == 0 {
		return
	}

	switch unit {
	case "b":
		mem = n * 8
	case "kb":
		mem = n / 1024
	case "mb":
		mem = n / (1024 * 1024)
	case "gb":
		mem = n / (1024 * 1024 * 1024)
	default:
		mem = n
	}

	return
}

// FormatAllocs converts an allocation count to the specified unit.
// Supported units: "K" (thousands), "M" (millions), "B" (billions), "T" (trillions).
// Empty unit string returns the value unchanged.
// Returns the converted allocation count as a float64.
func FormatAllocs(n float64, unit string) (allocs float64) {
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

	return
}
