// Package charts is the registry and per-chart typed Config layer that
// replaces the legacy two-tier DatasetSettings/ChartSettings model. Each
// chart type self-registers in its own subpackage via init(); callers use
// Register, New, Decode, and Registered to work with configs.
//
// The ChartConfig interface and registry live in the shared package (so
// Dataset.Settings can reference the interface without creating an import
// cycle — per-chart packages each import shared for Sort, so the cycle would
// be shared→config/charts→bar→shared). This file is a thin re-export of the
// registry functions for callers that prefer the charts.* namespace.
package charts

import (
	"encoding/json"

	"github.com/goptics/vizb/shared"
)

// Factory produces a fresh zero-value ChartConfig of a registered type. Alias
// for shared.ChartConfigFactory; see chart_config.go for the canonical type.
type Factory = shared.ChartConfigFactory

// Register adds a chart type's factory to the registry. Thin wrapper around
// shared.RegisterChartConfig.
func Register(chartType string, f Factory) {
	shared.RegisterChartConfig(chartType, f)
}

// New constructs a fresh zero-value ChartConfig of the given type. Thin
// wrapper around shared.NewChartConfig.
func New(chartType string) (shared.ChartConfig, error) {
	return shared.NewChartConfig(chartType)
}

// Decode unmarshals raw JSON into a ChartConfig of the given type. Thin
// wrapper around shared.DecodeChartConfig.
func Decode(chartType string, raw json.RawMessage) (shared.ChartConfig, error) {
	return shared.DecodeChartConfig(chartType, raw)
}

// Registered returns the sorted list of all registered chart types. Thin
// wrapper around shared.RegisteredChartConfigs.
func Registered() []string {
	return shared.RegisteredChartConfigs()
}
