package shared

import (
	"fmt"
	"strings"
)

// ECharts built-in series symbols (case-insensitive on CLI).
var echartsBuiltinSymbols = map[string]struct{}{
	"circle": {}, "rect": {}, "roundrect": {}, "triangle": {},
	"diamond": {}, "pin": {}, "arrow": {}, "none": {},
}

// ValidateSymbol reports whether s is an ECharts-accepted series symbol:
// built-in name, image://, path://, or raw SVG path (starts with M/m).
func ValidateSymbol(s string) error {
	if s == "" {
		return nil
	}
	if _, ok := echartsBuiltinSymbols[strings.ToLower(s)]; ok {
		return nil
	}
	if strings.HasPrefix(s, "image://") || strings.HasPrefix(s, "path://") {
		return nil
	}
	if len(s) > 0 && (s[0] == 'M' || s[0] == 'm') {
		return nil
	}
	return fmt.Errorf(
		"unknown symbol %q; use ECharts built-ins (circle, rect, roundRect, triangle, diamond, pin, arrow, none) or image:// / path:// / SVG path",
		s,
	)
}

// ValidateSymbolSize reports whether size is a positive marker diameter.
func ValidateSymbolSize(size float64) error {
	if size <= 0 {
		return fmt.Errorf("symbol size must be greater than 0, got %g", size)
	}
	return nil
}
