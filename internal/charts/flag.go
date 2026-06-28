package charts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goptics/vizb/internal/flags"
)

// --- Base flags: every chart accepts these as --chart keys. ---

var (
	// SwapFlag's validation is axis-dependent, so it is performed by the
	// override parser (which holds the runtime axes), not here.
	SwapFlag = flags.Flag{Name: "swap", Usage: "Axis permutation override", Kind: flags.KindString, JSONKey: "swap"}

	SortFlag = flags.Flag{
		Name: "sort", Shorthand: "s", Usage: "Sort order (asc, desc)", Kind: flags.KindString, JSONKey: "sort",
		Validate: ValidateSortValue,
		Encode:   func(v any) any { return map[string]any{"enabled": true, "order": strings.ToLower(v.(string))} },
		Label:    "sort order", ValidSet: []string{"asc", "desc"}, Normalizer: strings.ToLower,
	}

	// LabelsFlag's cobra flag is --show-labels (-l), but its --chart override key
	// is the shorter "labels"; Key carries that divergence.
	LabelsFlag = flags.Flag{Name: "show-labels", Shorthand: "l", Key: "labels", Usage: "Show labels on charts", Kind: flags.KindBool, JSONKey: "showLabels"}

	StatFlag = flags.Flag{
		Name:    "stat",
		Usage:   "Compute statistics (all categories when bare; comma-delimited list for specific categories)",
		Kind:    flags.KindStat,
		JSONKey: "stat",
	}
)

// BaseChartFlags are the --chart keys valid for every chart type. Each chart's
// flag list is composed by prepending a clone of BaseChartFlags before the
// chart's own variable flags (declared in cmd/charts/<c>/<c>.go).
var BaseChartFlags = []flags.Flag{SwapFlag, SortFlag, LabelsFlag, StatFlag}

// --- Variable flags: composed by the charts that carry them. ---

var (
	ScaleFlag = flags.Flag{
		Name: "scale", Shorthand: "S", Usage: "Scale type (linear, log)",
		Kind: flags.KindString, Default: "linear", JSONKey: "scale",
		Validate:   ValidateScaleValue,
		Encode:     func(v any) any { return strings.ToLower(v.(string)) },
		Label:      "scale",
		ValidSet:   []string{"linear", "log"},
		Normalizer: strings.ToLower,
	}
	ThreeDFlag = flags.Flag{
		Name: "3d", Usage: "Enable value 3D for x+y data (y categories on depth, metric on height)",
		Kind: flags.KindBool, JSONKey: "threeD",
		Rule: []flags.RuleFn{RequiresAxes("x", "y")},
	}
	ThreeDRotateFlag = flags.Flag{
		Name: "3d-rotate", Usage: "Auto-rotate the 3D scene (only applies when z-axis data is present)",
		Kind: flags.KindBool, JSONKey: "threeDRotate",
		Rule: []flags.RuleFn{RequiresZAxis()},
	}
	ThreeDVisualMapFlag = flags.Flag{
		Name: "3d-visualmap", Usage: "Color 3D bars/lines by metric value (visualMap gradient)",
		Kind: flags.KindBool, JSONKey: "threeDVisualMap",
		Rule: []flags.RuleFn{Requires3DMode()},
	}
	VisualMapFlag = flags.Flag{
		Name: "visualmap", Usage: "Color 2D scatter points by metric (visualMap gradient)",
		Kind: flags.KindBool, JSONKey: "visualMap",
		Rule: []flags.RuleFn{OnlyScatter2D()},
	}
	SymbolFlag = flags.Flag{
		Name:     "symbol",
		Usage:    "Marker symbol (ECharts built-in: circle, rect, roundRect, triangle, diamond, pin, arrow, none; or path:// / image:// / SVG path)",
		Kind:     flags.KindString,
		JSONKey:  "symbol",
		Validate: ValidateSymbolValue,
	}
	SymbolSizeFlag = flags.Flag{
		Name:     "symbol-size",
		Usage:    "Marker size in pixels (overrides default sizing)",
		Kind:     flags.KindFloat,
		JSONKey:  "symbolSize",
		Validate: ValidateSymbolSizeValue,
	}
)

// --- Pure validators (no shared dependency) usable by descriptors. ---

// ValidateScaleValue reports whether s is a valid scale (linear/log),
// case-insensitively.
func ValidateScaleValue(s string) error {
	switch strings.ToLower(s) {
	case "linear", "log":
		return nil
	}
	return fmt.Errorf("scale value %q is invalid (must be \"linear\" or \"log\")", s)
}

// ValidateSortValue reports whether s is a valid sort order (asc/desc),
// case-insensitively.
func ValidateSortValue(s string) error {
	switch strings.ToLower(s) {
	case "asc", "desc":
		return nil
	}
	return fmt.Errorf("sort value %q is invalid (must be \"asc\" or \"desc\")", s)
}

// echartsBuiltinSymbols are the ECharts built-in series symbols (case-insensitive).
var echartsBuiltinSymbols = map[string]struct{}{
	"circle": {}, "rect": {}, "roundrect": {}, "triangle": {},
	"diamond": {}, "pin": {}, "arrow": {}, "none": {},
}

// ValidateSymbolValue reports whether s is an ECharts-accepted series symbol:
// built-in name, image://, path://, or raw SVG path (starts with M/m).
func ValidateSymbolValue(s string) error {
	if s == "" {
		return nil
	}
	if _, ok := echartsBuiltinSymbols[strings.ToLower(s)]; ok {
		return nil
	}
	if strings.HasPrefix(s, "image://") || strings.HasPrefix(s, "path://") {
		return nil
	}
	if s[0] == 'M' || s[0] == 'm' {
		return nil
	}
	return fmt.Errorf(
		"unknown symbol %q; use ECharts built-ins (circle, rect, roundRect, triangle, diamond, pin, arrow, none) or image:// / path:// / SVG path",
		s,
	)
}

// ValidateSymbolSizeValue reports whether s parses to a positive marker diameter.
func ValidateSymbolSizeValue(s string) error {
	size, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("symbol size %q must be a number", s)
	}
	if size <= 0 {
		return fmt.Errorf("symbol size must be greater than 0, got %g", size)
	}
	return nil
}
