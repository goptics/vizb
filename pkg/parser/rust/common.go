package rust

import (
	"regexp"
	"strconv"
	"strings"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func toNsMultiplier(unit string) float64 {
	switch unit {
	case "ns":
		return 1
	case "µs", "μs":
		return 1e3
	case "ms":
		return 1e6
	case "s":
		return 1e9
	}
	return 1
}

func parseNum(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseFloat(s, 64)
	return n
}
