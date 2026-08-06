package parser

import "slices"

// ResolveMode determines the parse Mode from the resolved Config. Call once
// after Select/SelectViews/Axes are populated (in ParseConfig) so downstream
// code switches on cfg.Mode instead of re-deriving it from predicates.
// When ChartTypes is known (set before parse in the pipeline), re-call so
// sankey solo --select resolves to ModeEdge.
//
// Resolution order:
//  1. Explicit grouping + Select → ModeGrouped
//  2. Solo SelectViews (no explicit grouping):
//     a. sankey in ChartTypes → ModeEdge (min 3 columns; validated separately)
//     b. len > 1 → ModeMultiStat (validated to 2-col dim,metric)
//     c. len == 1 → ModeValue or ModeMixed (caller resolves after type inference)
//  3. Otherwise → ModeAuto
//
// Mixed vs value for a solo single view is not known until ResolveAxesTypes
// runs (it needs the data). ResolveMode sets ModeValue for a single solo view;
// the parser sets ModeMixed on its local cfg copy after type inference. The
// dataset builder treats both identically for axes derivation.
func ResolveMode(cfg Config) Mode {
	if IsExplicitGrouping(cfg) && len(cfg.Select) > 0 {
		return ModeGrouped
	}
	if len(cfg.SelectViews) > 0 && !IsExplicitGrouping(cfg) {
		if HasSankeyChart(cfg) {
			return ModeEdge
		}
		if len(cfg.SelectViews) > 1 {
			return ModeMultiStat
		}
		return ModeValue
	}
	return ModeAuto
}

// HasSankeyChart reports whether any requested chart type is sankey.
func HasSankeyChart(cfg Config) bool {
	return slices.Contains(cfg.ChartTypes, "sankey")
}

// IsGrouped reports whether cfg is in grouped stat-column mode.
func (m Mode) IsGrouped() bool { return m == ModeGrouped }

// IsSelectAxis reports whether cfg is solo --select axis mode
// (value, mixed, multi-stat, or edge).
func (m Mode) IsSelectAxis() bool {
	return m == ModeValue || m == ModeMixed || m == ModeMultiStat || m == ModeEdge
}

// IsMultiStat reports multi-stat solo --select mode.
func (m Mode) IsMultiStat() bool { return m == ModeMultiStat }

// IsEdge reports sankey edge-list solo --select mode.
func (m Mode) IsEdge() bool { return m == ModeEdge }
