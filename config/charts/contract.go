// Package charts is the registry and per-chart typed Config layer that
// replaces the legacy two-tier DatasetSettings/ChartSettings model. Each
// chart type self-registers in its own subpackage via init(); callers use
// Register, New, Decode, and Registered to work with configs.
package charts

// ChartConfig is the tiny contract every per-chart config implements. The
// interface is intentionally minimal — validation and defaulting live in
// each config package, co-located with the data the rules apply to.
type ChartConfig interface {
	ChartType() string
}
