package charts

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/goptics/vizb/internal/flags"
)

// Factory produces a fresh zero-value ChartConfig of a registered type. Used
// by Decode to build a typed instance before unmarshalling JSON into it.
type Factory func() ChartConfig

// Spec is the factory for a chart type: how to build its config. Cobra
// metadata (Use/Short/Long) lives in cmd/cli.ChartMeta; flag descriptors
// are stored separately via SetFlags. Each chart subpackage registers one
// Spec via cmd/charts/<c> in init().
type Spec struct {
	Type    string  // discriminator and registry key, e.g. "bar"
	Factory Factory // builds a fresh zero-value Config
}

// registeredFlags maps a chart type to its variable flag descriptors. Set by
// SetFlags from cmd/charts/<c> init() and read by FlagsFor / AllFlagNames.
var registeredFlags = map[string][]flags.Flag{}

// SetFlags stores the per-chart flag descriptors for chartType. It must be
// called before FlagsFor / AllFlagNames are consumed (it is, from init()).
func SetFlags(chartType string, ff []flags.Flag) {
	registeredFlags[chartType] = ff
}

// registry maps a chart type to its Spec. Populated by per-chart init() calls.
var registry = map[string]Spec{}

// Register adds a chart type's Spec to the registry. Panics on duplicate
// registration to surface a programming error at startup, before the registry
// is used.
func Register(s Spec) {
	if _, exists := registry[s.Type]; exists {
		panic(fmt.Sprintf("charts: chart type %q already registered", s.Type))
	}
	registry[s.Type] = s
}

// Get returns the Spec for a chart type and whether it is registered.
func Get(chartType string) (Spec, bool) {
	s, ok := registry[chartType]
	return s, ok
}

// New constructs a fresh zero-value ChartConfig of the given type. Returns
// an error if the type is not registered.
func New(chartType string) (ChartConfig, error) {
	s, ok := registry[chartType]
	if !ok {
		return nil, fmt.Errorf("unknown chart type %q (registered: %v)", chartType, Registered())
	}
	return s.Factory(), nil
}

// Decode unmarshals raw JSON into a ChartConfig of the given type. The
// factory in the registry provides a fresh zero value with the right concrete
// type; we take its address and let json.Unmarshal populate it.
func Decode(chartType string, raw json.RawMessage) (ChartConfig, error) {
	s, ok := registry[chartType]
	if !ok {
		return nil, fmt.Errorf("unknown chart type %q (registered: %v)", chartType, Registered())
	}
	cfg := s.Factory()
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

// Specs returns every registered Spec, sorted by type, for building cobra
// subcommands.
func Specs() []Spec {
	out := make([]Spec, 0, len(registry))
	for _, t := range Registered() {
		out = append(out, registry[t])
	}
	return out
}

// FlagsFor returns the full --chart key set for a chart type. Empty for an
// unregistered type.
func FlagsFor(chartType string) []flags.Flag {
	return registeredFlags[chartType]
}

// AllFlagNames returns the set of every flag name registered by any chart.
// Used to tell a cross-chart key (valid elsewhere, dropped here) from a typo
// (valid nowhere).
func AllFlagNames() map[string]bool {
	names := map[string]bool{}
	for _, ff := range registeredFlags {
		for _, f := range ff {
			names[f.EffectiveKey()] = true
		}
	}
	return names
}
