package charts

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Factory produces a fresh zero-value ChartConfig of a registered type. Used
// by Decode to build a typed instance before unmarshalling JSON into it.
type Factory func() ChartConfig

// registry maps a chart type to a Factory that builds a fresh zero value of
// its Config struct. Populated by per-chart init() calls.
var registry = map[string]Factory{}

// Register adds a chart type's factory to the registry. Panics on duplicate
// registration to surface a programming error at startup, before the registry
// is used.
func Register(chartType string, f Factory) {
	if _, exists := registry[chartType]; exists {
		panic(fmt.Sprintf("charts: chart type %q already registered", chartType))
	}
	registry[chartType] = f
}

// New constructs a fresh zero-value ChartConfig of the given type. Returns
// an error if the type is not registered.
func New(chartType string) (ChartConfig, error) {
	f, ok := registry[chartType]
	if !ok {
		return nil, fmt.Errorf("unknown chart type %q (registered: %v)", chartType, Registered())
	}
	return f(), nil
}

// Decode unmarshals raw JSON into a ChartConfig of the given type. The
// factory in the registry provides a fresh zero value with the right concrete
// type; we take its address and let json.Unmarshal populate it.
func Decode(chartType string, raw json.RawMessage) (ChartConfig, error) {
	f, ok := registry[chartType]
	if !ok {
		return nil, fmt.Errorf("unknown chart type %q (registered: %v)", chartType, Registered())
	}
	cfg := f()
	if err := json.Unmarshal(raw, cfg); err != nil {
		return nil, fmt.Errorf("decode %s config: %w", chartType, err)
	}
	return cfg, nil
}

// Registered returns the sorted list of all registered chart types. Useful
// for validation, --chart help text, and tests.
func Registered() []string {
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
