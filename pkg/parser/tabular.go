package parser

import (
	"fmt"
	"strconv"

	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
)

// RowReader abstracts one input row for the shared tabular parse functions.
// CSV and JSON each provide an adapter; the shared parseMixedMode/parseValueMode/
// parseSelectStatMode implementations read cells through this interface so the
// ~250 lines of CSV↔JSON duplication collapse into one implementation each.
type RowReader interface {
	// Cell returns the raw string value of the named column on the current row.
	// present is false when the column is absent or the row is too short.
	Cell(source string) (raw string, present bool)
	// Numeric returns the finite float64 value of the named column, or false
	// when absent/empty/non-finite.
	Numeric(source string) (float64, bool)
	// AvailableColumns returns the column/field names in file order (for errors).
	AvailableColumns() []string
	// FlagLabel is "--select" or "--axes" — the flag name shown in error messages.
	FlagLabel() string
}

// ParseMixedMode maps one categorical column to x and numeric columns to y[,z];
// each row becomes a point with empty stats (no aggregation). cfg.Axes must have
// AxisType set (category/value) from ResolveAxesTypes.
func ParseMixedMode(rows []RowReader, cfg Config) []shared.DataPoint {
	type slot struct{ kind string }
	slots := make(map[string]slot, len(cfg.Axes))
	for _, spec := range cfg.Axes {
		slots[spec.AxisKey] = slot{kind: spec.AxisType}
	}

	var results []shared.DataPoint
	for _, row := range rows {
		var dp shared.DataPoint
		complete := true
		for key, sl := range slots {
			if sl.kind == "category" {
				cell, ok := row.Cell(specSourceForKey(cfg, key))
				if !ok || cell == "" {
					complete = false
					break
				}
				assignAxis(&dp, key, cell)
				continue
			}
			v, ok := row.Numeric(specSourceForKey(cfg, key))
			if !ok {
				complete = false
				break
			}
			formatted := strconv.FormatFloat(utils.FormatNumber(v, cfg.NumberUnit), 'g', -1, 64)
			assignAxis(&dp, key, formatted)
		}
		if !complete {
			continue
		}
		results = append(results, dp)
	}
	return results
}

// ParseValueMode implements value mode: each named numeric column becomes a
// coordinate on x, y[, z]; each row becomes a raw point with no stat series.
// A missing or fully non-numeric axis column is fatal (rejected upstream by
// ResolveAxesTypes via the kindFn); a missing/non-numeric metric column is
// fatal (validated here); an individual row missing a finite cell is skipped.
func ParseValueMode(rows []RowReader, cfg Config) []shared.DataPoint {
	results, err := ParseValueModeE(rows, cfg)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
	return results
}

// ParseValueModeE is the error-returning variant used by reader-based parsers.
func ParseValueModeE(rows []RowReader, cfg Config) ([]shared.DataPoint, error) {
	if cfg.MetricColumn != "" && len(rows) > 0 {
		present, anyNum := false, false
		for _, row := range rows {
			if _, ok := row.Cell(cfg.MetricColumn); ok {
				present = true
			}
			if _, ok := row.Numeric(cfg.MetricColumn); ok {
				anyNum = true
				break
			}
		}
		if !present {
			return nil, fmt.Errorf("%s metric column %q not found; available: %v", rows[0].FlagLabel(), cfg.MetricColumn, rows[0].AvailableColumns())
		}
		if !anyNum {
			return nil, fmt.Errorf("%s metric column %q is not numeric", rows[0].FlagLabel(), cfg.MetricColumn)
		}
	}

	var results []shared.DataPoint
	for _, row := range rows {
		var dp shared.DataPoint
		dst := []*string{&dp.XAxis, &dp.YAxis, &dp.ZAxis}
		complete := true
		for i, spec := range cfg.Axes {
			if i >= len(dst) {
				break
			}
			v, ok := row.Numeric(spec.Source)
			if !ok {
				complete = false
				break
			}
			*dst[i] = strconv.FormatFloat(utils.FormatNumber(v, cfg.NumberUnit), 'g', -1, 64)
		}
		if !complete {
			continue
		}
		if cfg.MetricColumn != "" {
			mv, ok := row.Numeric(cfg.MetricColumn)
			if !ok {
				continue
			}
			dp.Metric = strconv.FormatFloat(utils.FormatNumber(mv, cfg.NumberUnit), 'g', -1, 64)
		}
		results = append(results, dp)
	}
	return results, nil
}

// ParseSelectStatMode parses repeatable solo --select into one dataset. When
// every flag shares the same dimension column, each input row becomes one point
// with multiple stats; otherwise each (row × view) stays a separate point.
func ParseSelectStatMode(rows []RowReader, cfg Config) []shared.DataPoint {
	results, err := ParseSelectStatModeE(rows, cfg)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
	return results
}

// ParseSelectStatModeE is the error-returning multi-select implementation.
func ParseSelectStatModeE(rows []RowReader, cfg Config) ([]shared.DataPoint, error) {
	merge := MultiSelectSharedDim(cfg.SelectViews)
	var results []shared.DataPoint
	for _, row := range rows {
		AppendMultiSelectStatPoint(&results, cfg.SelectViews, cfg.NumberUnit, merge, func(view SelectView) (MultiSelectRowStat, bool) {
			if len(view.Columns) < 2 {
				return MultiSelectRowStat{}, false
			}
			dim, metric := view.Columns[0], view.Columns[1]
			dimVal, ok := row.Cell(dim.Source)
			if !ok || dimVal == "" {
				return MultiSelectRowStat{}, false
			}
			v, ok := row.Numeric(metric.Source)
			if !ok {
				return MultiSelectRowStat{}, false
			}
			return MultiSelectRowStat{DimVal: dimVal, Value: v}, true
		})
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no dataset found")
	}
	return results, nil
}

// DispatchSelectMode routes a solo --select (or --axes/auto-value) Config to the
// right parse function after running ResolveAxesTypes. Returns the parsed
// DataPoints. Called by the CSV/JSON entry points — they pass their RowReader
// slice and kindFn. The flag label is baked into kindFn by the caller.
func DispatchSelectMode(rows []RowReader, cfg *Config, kindFn AxisColumnKind) []shared.DataPoint {
	results, err := DispatchSelectModeE(rows, cfg, kindFn)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
	return results
}

// DispatchSelectModeE routes select/value parsing without process termination.
func DispatchSelectModeE(rows []RowReader, cfg *Config, kindFn AxisColumnKind) ([]shared.DataPoint, error) {
	if cfg.Mode.IsMultiStat() {
		return ParseSelectStatModeE(rows, *cfg)
	}
	axesCfg := SelectViewAxesCfg(*cfg)
	if err := ResolveAxesTypes(&axesCfg, kindFn); err != nil {
		return nil, err
	}
	// Propagate resolved AxisType back to the caller's SelectViews so
	// DatasetAxesForSelectView can pick MixedAxes vs ValueAxes without
	// re-inferring from DataPoint strings.
	if len(cfg.SelectViews) > 0 && len(axesCfg.Axes) == len(cfg.SelectViews[0].Columns) {
		for i := range axesCfg.Axes {
			cfg.SelectViews[0].Columns[i].AxisType = axesCfg.Axes[i].AxisType
		}
	}
	if isMixedAxes(axesCfg) {
		cfg.Mode = ModeMixed
		return ParseMixedMode(rows, axesCfg), nil
	}
	return ParseValueModeE(rows, axesCfg)
}

// isMixedAxes is the post-ResolveAxesTypes mixed check (replaces the old
// IsMixedMode predicate — operated on the local axesCfg copy).
func isMixedAxes(cfg Config) bool {
	hasCat, hasVal := false, false
	for _, s := range cfg.Axes {
		switch s.AxisType {
		case "category":
			hasCat = true
		case "value":
			hasVal = true
		}
	}
	return hasCat && hasVal
}

// IsMixedAxes reports whether the Axes carry one category + at least one value
// (post ResolveAxesTypes). Used by the dataset builder for the --axes mixed path.
func IsMixedAxes(cfg Config) bool { return isMixedAxes(cfg) }

func specSourceForKey(cfg Config, key string) string {
	for _, s := range cfg.Axes {
		if s.AxisKey == key {
			return s.Source
		}
	}
	return ""
}

func assignAxis(dp *shared.DataPoint, key, value string) {
	switch key {
	case "x":
		dp.XAxis = value
	case "y":
		dp.YAxis = value
	case "z":
		dp.ZAxis = value
	}
}
