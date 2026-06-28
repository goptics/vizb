package shared

import (
	"fmt"

	internal_charts "github.com/goptics/vizb/internal/charts"
)

// ValidateSymbol reports whether s is an ECharts-accepted series symbol. The
// canonical implementation lives in config/charts (shared-free so chart flag
// descriptors can reference it); this delegates to keep existing callers stable.
func ValidateSymbol(s string) error {
	return internal_charts.ValidateSymbolValue(s)
}

// ValidateSymbolSize reports whether size is a positive marker diameter.
func ValidateSymbolSize(size float64) error {
	if size <= 0 {
		return fmt.Errorf("symbol size must be greater than 0, got %g", size)
	}
	return nil
}
