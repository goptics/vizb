package charts

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/goptics/vizb/config/flags"
)

// Factory produces a fresh zero-value ChartConfig of a registered type. Used
// by Decode to build a typed instance before unmarshalling JSON into it.
type Factory func() ChartConfig

// Spec is everything the rest of the program needs to know about a chart type:
// how to build its config, its cobra command metadata, and its variable flag
// set. Each chart subpackage registers one Spec in init(), mirroring how
// pkg/parser parsers self-register.
type Spec struct {
	Type    string       // discriminator and registry key, e.g. "bar"
	Use     string       // cobra Use line, e.g. "bar [target]"
	Short   string       // cobra Short help
	Long    string       // cobra Long help
	Factory Factory      // builds a fresh zero-value Config
	Flags   []flags.Flag // per-chart flags; each chart's init prepends BaseChartFlags
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

// FlagsFor returns the full --chart key set for a chart type: the chart's own
// Flags slice (which each chart init prepends with BaseChartFlags). Empty for
// an unregistered type.
func FlagsFor(chartType string) []flags.Flag {
	s, ok := registry[chartType]
	if !ok {
		return nil
	}
	return s.Flags
}

// AllFlagNames returns the set of every flag name registered by any chart.
// Used to tell a cross-chart key (valid elsewhere, dropped here) from a typo
// (valid nowhere).
func AllFlagNames() map[string]bool {
	names := map[string]bool{}
	for _, s := range registry {
		for _, f := range s.Flags {
			names[f.EffectiveKey()] = true
		}
	}
	return names
}
