// Package bar defines the typed Config for bar charts.
package bar

import (
	"github.com/goptics/vizb/internal/charts"
	"github.com/goptics/vizb/shared"
)

// Type is the chart-type discriminator written to JSON and used as the
// registry key.
const Type = "bar"

// Config is the per-chart typed config for bar charts. bar/line are the only
// chart types that carry a Scale (linear/log) and ThreeDRotate (3D) — pie,
// heatmap, and radar omit them.
type Config struct {
	Type            string             `json:"type"` // always "bar"
	Swap            string             `json:"swap,omitempty"`
	Sort            *shared.Sort       `json:"sort,omitempty"`
	Scale           string             `json:"scale,omitempty"`
	Stack           *bool              `json:"stack,omitempty"`
	ShowLabels      *bool              `json:"showLabels,omitempty"`
	ThreeDRotate    *bool              `json:"threeDRotate,omitempty"`
	ThreeD          *bool              `json:"threeD,omitempty"`
	ThreeDVisualMap *bool              `json:"threeDVisualMap,omitempty"`
	Horizontal      *bool              `json:"horizontal,omitempty"`
	BorderRadius    *int               `json:"borderRadius,omitempty"` // Added for border radius support
	Stat            *shared.StatConfig `json:"stat,omitempty"`
}

// ChartType returns the chart-type discriminator; satisfies charts.ChartConfig.
func (Config) ChartType() string { return Type }

func (c Config) StatEnabled() bool  { return c.Stat.StatEnabled() }
func (c Config) StatMath() []string { return c.Stat.StatMath() }
func (c Config) SwapString() string { return c.Swap }

// New returns a fresh zero-value bar chart Config.
func New() charts.ChartConfig { return &Config{Type: Type} }

// --- Border Radius Helper Methods ---

// ShouldApplyBorderRadius returns true if border radius should be applied
func (c *Config) ShouldApplyBorderRadius() bool {
	return c.BorderRadius != nil && *c.BorderRadius > 0
}

// GetBorderRadius returns the border radius value or 0 if not set
func (c *Config) GetBorderRadius() int {
	if c.BorderRadius == nil {
		return 0
	}
	return *c.BorderRadius
}

// IsStacked returns true if stacking is enabled
func (c *Config) IsStacked() bool {
	return c.Stack != nil && *c.Stack
}

// IsHorizontal returns true if bars are horizontal
func (c *Config) IsHorizontal() bool {
	return c.Horizontal != nil && *c.Horizontal
}

// GetCornerValues returns the appropriate corner radius values based on configuration.
// Returns [topLeft, topRight, bottomRight, bottomLeft] for ECharts.
// This is used by the rendering system to apply the correct corners.
func (c *Config) GetCornerValues() [4]int {
	if !c.ShouldApplyBorderRadius() {
		return [4]int{0, 0, 0, 0}
	}

	radius := c.GetBorderRadius()
	isHorizontal := c.IsHorizontal()
	
	// For stacked bars, the rendering system should call this for each series
	// passing isTopSeries parameter separately. This method returns the values
	// for a single series assuming it's the top/free end.
	if isHorizontal {
		// Horizontal: free end is on the right
		// [topLeft, topRight, bottomRight, bottomLeft]
		return [4]int{0, radius, radius, 0}
	}
	// Vertical: free end is on the top
	return [4]int{radius, radius, 0, 0}
}

// GetCornerValuesForSeries returns corner values for a specific series in a stack.
// isTopSeries should be true only for the topmost series in a stacked chart.
func (c *Config) GetCornerValuesForSeries(isTopSeries bool) [4]int {
	if !c.ShouldApplyBorderRadius() {
		return [4]int{0, 0, 0, 0}
	}

	isStacked := c.IsStacked()
	
	// If not stacked, all series get the radius
	// If stacked, only the top series gets the radius
	if isStacked && !isTopSeries {
		// Lower segments in stacked bars - no rounding
		return [4]int{0, 0, 0, 0}
	}
	
	// Top series or non-stacked - apply the radius
	return c.GetCornerValues()
}