package utils

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

func FormatAllocs(n uint64, unit string) (allocs uint64) {
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
