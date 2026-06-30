package parser

// ResolveMode determines the parse Mode from the resolved Config. Call once
// after Select/SelectViews/Axes are populated (in ParseConfig) so downstream
// code switches on cfg.Mode instead of re-deriving it from predicates.
//
// Resolution order:
//  1. Explicit grouping + Select → ModeGrouped
//  2. Solo SelectViews (no explicit grouping):
//     a. len > 1 → ModeMultiStat (validated to 2-col dim,metric)
//     b. len == 1 → ModeValue or ModeMixed (caller resolves after type inference)
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
		if len(cfg.SelectViews) > 1 {
			return ModeMultiStat
		}
		return ModeValue
	}
	return ModeAuto
}

// IsGrouped reports whether cfg is in grouped stat-column mode.
func (m Mode) IsGrouped() bool { return m == ModeGrouped }

// IsSelectAxis reports whether cfg is solo --select axis mode (value, mixed, or multi-stat).
func (m Mode) IsSelectAxis() bool { return m == ModeValue || m == ModeMixed || m == ModeMultiStat }

// IsMultiStat reports multi-stat solo --select mode.
func (m Mode) IsMultiStat() bool { return m == ModeMultiStat }
