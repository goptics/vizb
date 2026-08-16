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
	SwapFlag = flags.Flag{Name: "swap", Usage: "Swap n/x/y/z axis assignment", Kind: flags.KindString, JSONKey: "swap"}

	SortFlag = flags.Flag{
		Name: "sort", Shorthand: "s", Usage: "Sort order (asc, desc)", Kind: flags.KindString, JSONKey: "sort",
		Validate: ValidateSortValue,
		Encode:   func(v any) any { return map[string]any{"enabled": true, "order": strings.ToLower(v.(string))} },
		Label:    "sort order", ValidSet: []string{"asc", "desc"}, Normalizer: strings.ToLower,
	}

	// LabelsFlag's cobra flag is --show-labels (-l), but its --chart override key
	// is the shorter "labels"; Key carries that divergence.
	LabelsFlag = flags.Flag{Name: "show-labels", Shorthand: "l", Key: "labels", Usage: "Show data labels on charts", Kind: flags.KindBool, JSONKey: "showLabels"}

	StatFlag = flags.Flag{
		Name:    "stat",
		Usage:   "Enable stats panel (all when bare; list categories to limit)",
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
		Name: "scale", Shorthand: "S", Usage: "Value scale (linear, log)",
		Kind: flags.KindString, Default: "linear", JSONKey: "scale",
		Validate:   ValidateScaleValue,
		Encode:     func(v any) any { return strings.ToLower(v.(string)) },
		Label:      "scale",
		ValidSet:   []string{"linear", "log"},
		Normalizer: strings.ToLower,
	}
	StackFlag = flags.Flag{
		Name: "stack", Usage: "Stack 2D grouped series",
		Kind:    flags.KindBool,
		JSONKey: "stack",
		Rule:    []flags.RuleFn{RequiresAxes("x", "y"), ExcludesAxes("z"), StackRequiresLinearScale()},
	}
	ThreeDFlag = flags.Flag{
		Name: "3d", Usage: "Force 3D for x+y data (y on depth, metric on height)",
		Kind: flags.KindBool, JSONKey: "threeD",
		Rule: []flags.RuleFn{RequiresAxes("x", "y")},
	}
	ThreeDRotateFlag = flags.Flag{
		Name: "3d-rotate", Usage: "Auto-rotate 3D scene (needs z-axis data)",
		Kind: flags.KindBool, JSONKey: "threeDRotate",
		Rule: []flags.RuleFn{RequiresAxes("z")},
	}
	ThreeDVisualMapFlag = flags.Flag{
		Name: "3d-visualmap", Usage: "Color 3D series by metric (visualMap)",
		Kind: flags.KindBool, JSONKey: "threeDVisualMap",
		Rule: []flags.RuleFn{Requires3DMode()},
	}
	VisualMapFlag = flags.Flag{
		Name: "visualmap", Usage: "Color 2D scatter points by metric (visualMap)",
		Kind: flags.KindBool, JSONKey: "visualMap",
		Rule: []flags.RuleFn{OnlyScatter2D()},
	}
	SymbolFlag = flags.Flag{
		Name:     "symbol",
		Usage:    "Marker symbol (circle, rect, diamond, ...; or path:// / image://)",
		Kind:     flags.KindString,
		JSONKey:  "symbol",
		Validate: ValidateSymbolValue,
	}
	SymbolSizeFlag = flags.Flag{
		Name:     "symbol-size",
		Usage:    "Marker size in pixels",
		Kind:     flags.KindFloat,
		JSONKey:  "symbolSize",
		Validate: ValidateSymbolSizeValue,
	}
	SmoothFlag = flags.Flag{
		Name:    "smooth",
		Usage:   "Smooth curved line segments",
		Kind:    flags.KindBool,
		JSONKey: "smooth",
	}
	HorizontalFlag = flags.Flag{
		Name:    "horizontal",
		Usage:   "Horizontal bars (categories on Y, values on X)",
		Kind:    flags.KindBool,
		JSONKey: "horizontal",
	}
	BorderRadiusFlag = flags.Flag{
		Name:       "border-radius",
		Usage:      "Corner radius in px: 1–4 non-negative integers (CSS/ECharts TL,TR,BR,BL). e.g. 8 or 8,8,0,0",
		Kind:       flags.KindString,
		JSONKey:    "borderRadius",
		MultiValue: true,
		Validate:   ValidateBorderRadiusValue,
		Encode:     EncodeBorderRadius,
	}
	// BgFlag is the bar-only category background: bare --bg turns it on, and a
	// semicolon bag of style fields (--bg color=…;borderColor=#000) adds typed
	// props. Encode injects the implicit "active": true on-switch; "active" is
	// not a user field. 2D only — the rule skips it on 3D bars.
	BgFlag = flags.Flag{
		Name: "bg",
		Usage: "Category background behind bars (2D only; bare = on, or style props " +
			"semicolon-separated: color, borderColor, borderWidth, borderType, borderRadius, " +
			"shadowBlur, shadowColor, shadowOffsetX, shadowOffsetY, opacity)",
		Kind:         flags.KindObject,
		JSONKey:      "background",
		Encode:       EncodeBgObject,
		ObjectFields: bgObjectFields(),
		Rule:         []flags.RuleFn{Excludes3DMode()},
	}
)

// bgObjectFields lists the typed style fields accepted inside the --bg object.
func bgObjectFields() []flags.ObjectField {
	return []flags.ObjectField{
		{Name: "color", Kind: flags.KindString},
		{Name: "borderColor", Kind: flags.KindString},
		{Name: "borderWidth", Kind: flags.KindFloat, Validate: ValidateNonNegativeNumberValue, Encode: EncodeNumber},
		{Name: "borderType", Kind: flags.KindString, Validate: ValidateBorderTypeValue},
		{Name: "borderRadius", Kind: flags.KindString, Validate: ValidateBorderRadiusValue, Encode: EncodeBorderRadius},
		{Name: "shadowBlur", Kind: flags.KindFloat, Validate: ValidateNonNegativeNumberValue, Encode: EncodeNumber},
		{Name: "shadowColor", Kind: flags.KindString},
		{Name: "shadowOffsetX", Kind: flags.KindFloat, Validate: ValidateNumberValue, Encode: EncodeNumber},
		{Name: "shadowOffsetY", Kind: flags.KindFloat, Validate: ValidateNumberValue, Encode: EncodeNumber},
		{Name: "opacity", Kind: flags.KindFloat, Validate: ValidateOpacityValue, Encode: EncodeNumber},
	}
}

// EncodeBgObject injects the implicit on-switch into a parsed --bg object bag
// payload (bare --bg → empty bag → {"active": true}).
func EncodeBgObject(v any) any {
	bag, _ := v.(map[string]any)
	if bag == nil {
		bag = map[string]any{}
	}
	bag["active"] = true
	return bag
}

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

// parseBorderRadius parses a comma-separated list of 1–4 non-negative integers
// (CSS/ECharts corner order: TL, TR, BR, BL).
func parseBorderRadius(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("border radius must be 1–4 non-negative integers (e.g. 8 or 8,8,0,0)")
	}
	parts := strings.Split(s, ",")
	if len(parts) > 4 {
		return nil, fmt.Errorf("border radius accepts at most 4 values, got %d", len(parts))
	}
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, fmt.Errorf("border radius %q must be an integer", s)
		}
		r, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("border radius %q must be an integer", p)
		}
		if r < 0 {
			return nil, fmt.Errorf("border radius must be non-negative (>= 0), got %d", r)
		}
		out = append(out, r)
	}
	return out, nil
}

// ValidateBorderRadiusValue reports whether s is 1–4 comma-separated non-negative integers.
func ValidateBorderRadiusValue(s string) error {
	_, err := parseBorderRadius(s)
	return err
}

// EncodeBorderRadius maps a validated CLI/string value to a []int seed payload
// (always an array; ECharts treats [8] as all corners).
func EncodeBorderRadius(v any) any {
	s, ok := v.(string)
	if !ok {
		return v
	}
	vals, err := parseBorderRadius(s)
	if err != nil {
		return v
	}
	return vals
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

// ValidateNumberValue reports whether s parses to a finite number (object-flag
// field validator, e.g. --bg shadowOffsetX).
func ValidateNumberValue(s string) error {
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return fmt.Errorf("value %q must be a number", s)
	}
	return nil
}

// ValidateNonNegativeNumberValue reports whether s parses to a number >= 0
// (object-flag field validator, e.g. --bg borderWidth).
func ValidateNonNegativeNumberValue(s string) error {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("value %q must be a number", s)
	}
	if n < 0 {
		return fmt.Errorf("value must be non-negative (>= 0), got %g", n)
	}
	return nil
}

// ValidateOpacityValue reports whether s parses to a number in 0..1
// (object-flag field validator, e.g. --bg opacity).
func ValidateOpacityValue(s string) error {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("opacity %q must be a number", s)
	}
	if n < 0 || n > 1 {
		return fmt.Errorf("opacity must be between 0 and 1, got %g", n)
	}
	return nil
}

// ValidateBorderTypeValue reports whether s is a valid background border type
// (solid, dashed, or dotted).
func ValidateBorderTypeValue(s string) error {
	switch s {
	case "solid", "dashed", "dotted":
		return nil
	}
	return fmt.Errorf("border type %q is invalid (must be solid, dashed, or dotted)", s)
}

// EncodeNumber maps a validated numeric object-flag field value to a float64
// payload (non-string input passes through unchanged).
func EncodeNumber(v any) any {
	s, ok := v.(string)
	if !ok {
		return v
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return v
	}
	return n
}
