package shared

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ChartConfig is the tiny contract every per-chart config implements. The
// interface is declared in shared (not in config/charts) so Dataset.Settings
// can be []ChartConfig without creating an import cycle: per-chart packages
// register themselves here, and config/charts re-exports the registry symbols
// for backward compatibility with its existing API.
type ChartConfig interface {
	ChartType() string
}

// ChartConfigFactory produces a fresh zero-value ChartConfig of a registered
// type. The registry's Decode calls the factory to get a typed instance, then
// unmarshals the JSON into it via reflection on the interface's concrete type.
type ChartConfigFactory func() ChartConfig

// chartConfigRegistry maps a chart type to its factory. Populated by
// per-chart init() calls (e.g. config/charts/bar/bar.go).
var chartConfigRegistry = map[string]ChartConfigFactory{}

// RegisterChartConfig adds a chart type's factory to the registry. Panics on
// duplicate registration to surface a programming error at startup, before
// the registry is used.
func RegisterChartConfig(chartType string, f ChartConfigFactory) {
	if _, exists := chartConfigRegistry[chartType]; exists {
		panic(fmt.Sprintf("shared: chart type %q already registered", chartType))
	}
	chartConfigRegistry[chartType] = f
}

// NewChartConfig constructs a fresh zero-value ChartConfig of the given type.
// Returns an error when the type is not registered.
func NewChartConfig(chartType string) (ChartConfig, error) {
	f, ok := chartConfigRegistry[chartType]
	if !ok {
		return nil, fmt.Errorf("unknown chart type %q (registered: %v)", chartType, RegisteredChartConfigs())
	}
	return f(), nil
}

// DecodeChartConfig unmarshals raw JSON into a ChartConfig of the given type.
// The factory in the registry provides a fresh zero value with the right
// concrete type; we take its return and let json.Unmarshal populate it.
func DecodeChartConfig(chartType string, raw json.RawMessage) (ChartConfig, error) {
	f, ok := chartConfigRegistry[chartType]
	if !ok {
		return nil, fmt.Errorf("unknown chart type %q (registered: %v)", chartType, RegisteredChartConfigs())
	}
	cfg := f()
	if err := json.Unmarshal(raw, cfg); err != nil {
		return nil, fmt.Errorf("decode %s config: %w", chartType, err)
	}
	return cfg, nil
}

// RegisteredChartConfigs returns the sorted list of all registered chart types.
// Used for validation, --chart help text, and tests.
func RegisteredChartConfigs() []string {
	out := make([]string, 0, len(chartConfigRegistry))
	for k := range chartConfigRegistry {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
